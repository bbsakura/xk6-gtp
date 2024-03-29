// cf. https://github.com/wmnsk/go-gtp/blob/master/examples/pgw/pgw.go
package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/ie"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

func getSubscriberIP(sub *gtpv2.Subscriber) (string, error) {
	subIPMap := map[string]string{
		"123451234567891": "10.10.10.1",
		"123451234567892": "10.10.10.2",
		"123451234567893": "10.10.10.3",
		"123451234567894": "10.10.10.4",
		"123451234567895": "10.10.10.5",
	}

	if ip, ok := subIPMap[sub.IMSI]; ok {
		return ip, nil
	}
	return "", fmt.Errorf("subscriber %s not found", sub.IMSI)
}

var (
	loggerCh = make(chan string)
	errCh    = make(chan error)
)

func handleEchoRequest(c *gtpv2.Conn, senderAddr net.Addr, msg message.Message) error {
	// this should never happen, as the type should have been assured by
	// msgHandlerMap before this function is called.
	if _, ok := msg.(*message.EchoRequest); !ok {
		return &gtpv2.UnexpectedTypeError{Msg: msg}
	}
	log.Printf("senderAddr: %s,received: %v\n", senderAddr, msg)

	// respond with EchoResponse.
	return c.RespondTo(
		senderAddr, msg, message.NewEchoResponse(0, ie.NewRecovery(c.RestartCounter)),
	)
}

func handleCreateSessionRequest(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
	loggerCh <- fmt.Sprintf("Received %s from %s", msg.MessageTypeName(), sgwAddr)

	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	csReqFromSGW := msg.(*message.CreateSessionRequest)

	// keep session information retrieved from the message.
	session := gtpv2.NewSession(sgwAddr, &gtpv2.Subscriber{Location: &gtpv2.Location{}})
	bearer := session.GetDefaultBearer()
	var err error
	if imsiIE := csReqFromSGW.IMSI; imsiIE != nil {
		imsi, err := imsiIE.IMSI()
		if err != nil {
			return err
		}
		session.IMSI = imsi

		// remove previous session for the same subscriber if exists.
		sess, err := c.GetSessionByIMSI(imsi)
		if err != nil {
			switch err.(type) {
			case *gtpv2.UnknownIMSIError:
				// whole new session. just ignore.
			default:
				return fmt.Errorf("got something unexpected: %w", err)
			}
		} else {
			c.RemoveSession(sess)
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.IMSI}
	}
	if msisdnIE := csReqFromSGW.MSISDN; msisdnIE != nil {
		session.MSISDN, err = msisdnIE.MSISDN()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.MSISDN}
	}
	if meiIE := csReqFromSGW.MEI; meiIE != nil {
		session.IMEI, err = meiIE.MobileEquipmentIdentity()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.MobileEquipmentIdentity}
	}
	if apnIE := csReqFromSGW.APN; apnIE != nil {
		bearer.APN, err = apnIE.AccessPointName()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.AccessPointName}
	}
	if netIE := csReqFromSGW.ServingNetwork; netIE != nil {
		session.MNC, err = netIE.MNC()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.ServingNetwork}
	}
	if ratIE := csReqFromSGW.RATType; ratIE != nil {
		session.RATType, err = ratIE.RATType()
		if err != nil {
			return err
		}
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.RATType}
	}

	// SGW TEID cplane
	if fteidcIE := csReqFromSGW.SenderFTEIDC; fteidcIE != nil {
		teid, err := fteidcIE.TEID()
		if err != nil {
			return err
		}
		session.AddTEID(gtpv2.IFTypeS5S8SGWGTPC, teid)
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.FullyQualifiedTEID}
	}

	// var teidOut uint32
	if brCtxIE := csReqFromSGW.BearerContextsToBeCreated; brCtxIE != nil {
		for _, childIE := range brCtxIE[0].ChildIEs {
			switch childIE.Type {
			case ie.EPSBearerID:
				// EBI: 2個め以降のbearerを取ることができる.例えば音声用とか
				bearer.EBI, err = childIE.EPSBearerID()
				if err != nil {
					return err
				}

			case ie.FullyQualifiedTEID: // sgw teid uplane
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

	bearer.SubscriberIP, err = getSubscriberIP(session.Subscriber)
	if err != nil {
		return err
	}

	cIP := strings.Split(c.LocalAddr().String(), ":")[0]
	uIP := strings.Split(*s5u, ":")[0]

	// PGW Cplane TEID
	var s5cFTEID *ie.IE
	if session.Subscriber.IMEI == "123451234567895" { // testing only
		s5cFTEID = ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8PGWGTPC, 111, uIP, "").WithInstance(1)
	} else {
		s5cFTEID = c.NewSenderFTEID(cIP, "").WithInstance(1)
	}
	// PGW Uplane TEID
	s5uFTEID := ie.NewFullyQualifiedTEID(gtpv2.IFTypeS5S8PGWGTPU, generateRandomUint32(), uIP, "")
	// SGW Cplane TEID
	s5sgwTEID, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
	if err != nil {
		return err
	}
	csRspFromPGW := message.NewCreateSessionResponse(
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
	if csReqFromSGW.SGWFQCSID != nil {
		csRspFromPGW.PGWFQCSID = ie.NewFullyQualifiedCSID(cIP, 1)
	}
	session.AddTEID(gtpv2.IFTypeS5S8PGWGTPC, s5cFTEID.MustTEID())
	session.AddTEID(gtpv2.IFTypeS5S8PGWGTPU, s5uFTEID.MustTEID())

	if err := c.RespondTo(sgwAddr, csReqFromSGW, csRspFromPGW); err != nil {
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

	loggerCh <- fmt.Sprintf("Session created with S-GW for subscriber: %s;\n\tS5C S-GW: %s, TEID->: %#x, TEID<-: %#x",
		session.Subscriber.IMSI, sgwAddr, s5sgwTEID, s5pgwTEID,
	)
	return nil
}

func handleDeleteSessionRequest(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
	loggerCh <- fmt.Sprintf("Received %s from %s", msg.MessageTypeName(), sgwAddr)

	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	session, err := c.GetSessionByTEID(msg.TEID(), sgwAddr)
	if err != nil {
		dsr := message.NewDeleteSessionResponse(
			0, 0,
			ie.NewCause(gtpv2.CauseIMSIIMEINotKnown, 0, 0, 0, nil),
		)
		if err := c.RespondTo(sgwAddr, msg, dsr); err != nil {
			return err
		}

		return err
	}

	// respond to S-GW with DeleteSessionResponse.
	teid, err := session.GetTEID(gtpv2.IFTypeS5S8SGWGTPC)
	if err != nil {
		loggerCh <- fmt.Sprintf("Error: %v", err)
		return nil
	}
	dsr := message.NewDeleteSessionResponse(
		teid, 0,
		ie.NewCause(gtpv2.CauseRequestAccepted, 0, 0, 0, nil),
	)
	if err := c.RespondTo(sgwAddr, msg, dsr); err != nil {
		return err
	}

	loggerCh <- fmt.Sprintf("Session deleted for Subscriber: %s", session.IMSI)
	c.RemoveSession(session)
	return nil
}

// Modify Request/Response in TAU with S-GW relocation
func handleModifyBearerRequest(c *gtpv2.Conn, sgwAddr net.Addr, msg message.Message) error {
	// SGW TEID cplane
	reqFromSGW := msg.(*message.ModifyBearerRequest)
	s5sgwTEID := uint32(0)
	if fteidcIE := reqFromSGW.SenderFTEIDC; fteidcIE != nil {
		teid, err := fteidcIE.TEID()
		if err != nil {
			return err
		}
		s5sgwTEID = teid
	} else {
		return &gtpv2.RequiredIEMissingError{Type: ie.FullyQualifiedTEID}
	}

	seqn := reqFromSGW.Header.SequenceNumber

	// cause: context not found
	// TODO...

	// cause: RequestAccepted
	rspFromPGW := message.NewModifyBearerResponse(
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
	if err := c.RespondTo(sgwAddr, reqFromSGW, rspFromPGW); err != nil {
		return err
	}

	return nil
}

func generateRandomUint32() uint32 {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return 0
	}

	return binary.BigEndian.Uint32(b)
}
