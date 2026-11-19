// Package refpgw provides the reference PGW (P-GW) handler set used both by
// the standalone binary in cmd/pgw and by tests that need a peer to drive
// against. Extracted from cmd/pgw so the handler responsibilities live in a
// reusable module and the CLI wrapper stays thin.
package refpgw

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

// Config configures the reference PGW handlers.
type Config struct {
	// S5UAddr is the local address (host[:port]) reported as the PGW U-plane
	// FTEID in the Create Session Response. Only the host portion is used.
	S5UAddr string
	// Logger receives structured log records for every handled message. When
	// nil, slog.Default() is used.
	Logger *slog.Logger
	// Subscribers maps IMSI to allocated PDN address. When nil, a small
	// built-in list of dummy subscribers (matching the pre-refactor
	// behaviour) is used.
	Subscribers map[string]string
}

// Handlers holds the reference PGW state (subscriber DB, logger, config) and
// exposes AddTo to install its handlers on a gtpv2.Conn.
type Handlers struct {
	cfg Config
}

// New constructs a Handlers value with sensible defaults for logger and
// subscribers.
func New(cfg Config) *Handlers {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Subscribers == nil {
		cfg.Subscribers = defaultSubscribers()
	}
	if cfg.S5UAddr == "" {
		cfg.S5UAddr = "127.0.0.1"
	}
	return &Handlers{cfg: cfg}
}

// AddTo registers the reference PGW's message handlers on conn. Overwrites
// any existing handlers for the same message types.
func (h *Handlers) AddTo(conn *gtpv2.Conn) {
	conn.AddHandlers(map[uint8]gtpv2.HandlerFunc{
		message.MsgTypeCreateSessionRequest: h.handleCreateSessionRequest,
		message.MsgTypeDeleteSessionRequest: h.handleDeleteSessionRequest,
		message.MsgTypeModifyBearerRequest:  h.handleModifyBearerRequest,
		message.MsgTypeEchoRequest:          h.handleEchoRequest,
	})
}

// defaultSubscribers is the fixed dummy IMSI→PDN-address map that the
// reference PGW has accepted since the project's smoke tests were written.
func defaultSubscribers() map[string]string {
	return map[string]string{
		"123451234567891": "10.10.10.1",
		"123451234567892": "10.10.10.2",
		"123451234567893": "10.10.10.3",
		"123451234567894": "10.10.10.4",
		"123451234567895": "10.10.10.5",
	}
}

func (h *Handlers) getSubscriberIP(sub *gtpv2.Subscriber) (string, error) {
	if ip, ok := h.cfg.Subscribers[sub.IMSI]; ok {
		return ip, nil
	}
	return "", fmt.Errorf("subscriber %q not found", sub.IMSI)
}

func (h *Handlers) handleEchoRequest(c *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
	if _, ok := msg.(*message.EchoRequest); !ok {
		return &gtpv2.UnexpectedTypeError{Msg: msg}
	}
	h.cfg.Logger.Info("received echo request",
		"peer", senderAddr.String(),
		"seq", msg.Sequence(),
	)
	return c.RespondTo(
		senderAddr, msg, message.NewEchoResponse(0, ie.NewRecovery(c.RestartCounter)),
	)
}

// handleCreateSessionRequest mirrors the 29.274 IE walk directly; splitting
// it further would fragment the reader's mental model without simplifying it.
//
//nolint:gocyclo // handler intentionally mirrors 29.274 IE walk
func (h *Handlers) handleCreateSessionRequest(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
	logger := h.cfg.Logger.With("peer", sgwAddr.String(), "seq", msg.Sequence(), "msg_type", "create_session_request")
	logger.Info("received")

	csReq, ok := msg.(*message.CreateSessionRequest)
	if !ok {
		return &gtpv2.UnexpectedTypeError{Msg: msg}
	}

	session := gtpv2.NewSession(sgwAddr, &gtpv2.Subscriber{Location: &gtpv2.Location{}})
	bearer := session.GetDefaultBearer()
	var err error
	if imsiIE := csReq.IMSI; imsiIE != nil {
		imsi, err := imsiIE.IMSI()
		if err != nil {
			return err
		}
		session.IMSI = imsi
		logger = logger.With("imsi", imsi)

		sess, err := c.GetSessionByIMSI(imsi)
		if err != nil {
			var unknown *gtpv2.UnknownIMSIError
			if !errors.As(err, &unknown) {
				return fmt.Errorf("look up session for IMSI %q: %w", imsi, err)
			}
		} else {
			c.RemoveSession(sess)
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.IMSI}
	}
	if msisdnIE := csReq.MSISDN; msisdnIE != nil {
		session.MSISDN, err = msisdnIE.MSISDN()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.MSISDN}
	}
	if meiIE := csReq.MEI; meiIE != nil {
		session.IMEI, err = meiIE.MobileEquipmentIdentity()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.MobileEquipmentIdentity}
	}
	if apnIE := csReq.APN; apnIE != nil {
		bearer.APN, err = apnIE.AccessPointName()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.AccessPointName}
	}
	if netIE := csReq.ServingNetwork; netIE != nil {
		session.MNC, err = netIE.MNC()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.ServingNetwork}
	}
	if ratIE := csReq.RATType; ratIE != nil {
		session.RATType, err = ratIE.RATType()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.RATType}
	}

	if fteidcIE := csReq.SenderFTEIDC; fteidcIE != nil {
		teid, err := fteidcIE.TEID()
		if err != nil {
			return err
		}
		session.AddTEID(gtpv2.IFTypeS5S8SGWGTPC, teid)
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.FullyQualifiedTEID}
	}

	if brCtxIE := csReq.BearerContextsToBeCreated; brCtxIE != nil {
		for _, childIE := range brCtxIE[0].ChildIEs {
			switch childIE.Type {
			case ie.EPSBearerID:
				bearer.EBI, err = childIE.EPSBearerID()
				if err != nil {
					return err
				}
			case ie.FullyQualifiedTEID:
				it, err := childIE.InterfaceType()
				if err != nil {
					return err
				}
				teidOut, err := childIE.TEID()
				if err != nil {
					return err
				}
				session.AddTEID(it, teidOut)
			}
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.BearerContext}
	}

	bearer.SubscriberIP, err = h.getSubscriberIP(session.Subscriber)
	if err != nil {
		logger.Warn("reject unknown subscriber", "error", err)
		return err
	}

	cIP := hostOf(c.LocalAddr().String())
	uIP := hostOf(h.cfg.S5UAddr)

	var s5cFTEID *ie.IE
	if session.IMEI == "123451234567895" { // testing only
		s5cFTEID = ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8PGWGTPC, 111, uIP, "").WithInstance(1)
	} else {
		s5cFTEID = c.NewSenderFTEID(cIP, "").WithInstance(1)
	}
	s5uFTEID := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8PGWGTPU, generateRandomUint32(), uIP, "")
	s5sgwTEID, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
	if err != nil {
		return err
	}
	csRsp := message.NewCreateSessionResponse(
		s5sgwTEID, 0,
		ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
		s5cFTEID,
		ie.NewPDNAddressAllocation(bearer.SubscriberIP),
		ie.NewAPNRestriction(gtpv2.APNRestrictionPublic2),
		ie.NewBearerContext(
			ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
			ie.NewEPSBearerID(bearer.EBI),
			s5uFTEID,
			ie.NewChargingID(bearer.ChargingID),
		),
	)
	if csReq.SGWFQCSID != nil {
		csRsp.PGWFQCSID = ie.NewFullyQualifiedCSID(cIP, 1)
	}
	session.AddTEID(gtpv2.IFTypeS5S8PGWGTPC, s5cFTEID.MustTEID())
	session.AddTEID(gtpv2.IFTypeS5S8PGWGTPU, s5uFTEID.MustTEID())

	if err := c.RespondTo(sgwAddr, csReq, csRsp); err != nil {
		return err
	}

	s5pgwTEID, err := session.GetTEID(gtpv2.IFTypeS5S8PGWGTPC)
	if err != nil {
		return err
	}
	c.RegisterSession(s5pgwTEID, session)
	if err := session.Activate(); err != nil {
		return err
	}

	logger.Info("session created",
		"sgw_teid", fmt.Sprintf("%#x", s5sgwTEID),
		"pgw_teid", fmt.Sprintf("%#x", s5pgwTEID),
	)
	return nil
}

func (h *Handlers) handleDeleteSessionRequest(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
	logger := h.cfg.Logger.With("peer", sgwAddr.String(), "seq", msg.Sequence(), "msg_type", "delete_session_request")
	logger.Info("received")

	session, err := c.GetSessionByTEID(msg.TEID(), sgwAddr)
	if err != nil {
		dsr := message.NewDeleteSessionResponse(
			0, 0,
			ie.NewCause(gtpv2.CauseIMSIIMEINotKnown, 0, 0, 0, nil),
		)
		if respErr := c.RespondTo(sgwAddr, msg, dsr); respErr != nil {
			return respErr
		}
		logger.Warn("unknown session, sent cause=IMSIIMEINotKnown", "error", err)
		return err
	}

	teid, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
	if err != nil {
		logger.Error("look up sgw teid", "error", err)
		return nil
	}
	dsr := message.NewDeleteSessionResponse(
		teid, 0,
		ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
	)
	if err := c.RespondTo(sgwAddr, msg, dsr); err != nil {
		return err
	}

	logger.Info("session deleted", "imsi", session.IMSI)
	c.RemoveSession(session)
	return nil
}

func (h *Handlers) handleModifyBearerRequest(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
	logger := h.cfg.Logger.With("peer", sgwAddr.String(), "seq", msg.Sequence(), "msg_type", "modify_bearer_request")

	req, ok := msg.(*message.ModifyBearerRequest)
	if !ok {
		return &gtpv2.UnexpectedTypeError{Msg: msg}
	}

	s5sgwTEID := uint32(0)
	if fteidcIE := req.SenderFTEIDC; fteidcIE != nil {
		teid, err := fteidcIE.TEID()
		if err != nil {
			return err
		}
		s5sgwTEID = teid
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.FullyQualifiedTEID}
	}

	seqn := req.SequenceNumber
	rsp := message.NewModifyBearerResponse(
		s5sgwTEID, seqn,
		ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
		ie.NewMSISDN("819010001000"),
		ie.NewBearerContext(
			ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
			ie.NewChargingID(0),
		),
		ie.NewRecovery(0),
		ie.NewAPNRestriction(gtpv2.APNRestrictionPublic2),
	)
	if err := c.RespondTo(sgwAddr, req, rsp); err != nil {
		return err
	}
	logger.Info("modify bearer answered", "sgw_teid", fmt.Sprintf("%#x", s5sgwTEID))
	return nil
}

func generateRandomUint32() uint32 {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return 0
	}
	return binary.BigEndian.Uint32(b)
}

// hostOf strips a trailing :port from a host[:port] string. Uses
// net.SplitHostPort so bracketed IPv6 addresses survive intact.
func hostOf(hostPort string) string {
	if host, _, err := net.SplitHostPort(hostPort); err == nil {
		return host
	}
	return strings.TrimSuffix(hostPort, ":")
}
