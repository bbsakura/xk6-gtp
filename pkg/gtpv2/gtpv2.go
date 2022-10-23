package gtpv2

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

type K6GTPv2 struct {
	Version string
	Conn    *gtpv2.Conn
	Ctx     context.Context
}

type ConnectionOptions struct {
	Saddr      string
	Daddr      string
	Count      int
	IFTypeName string
}

func (c *K6GTPv2) Connect(options ConnectionOptions) error {
	var err error
	saddr, err := net.ResolveIPAddr("ip", options.Saddr)
	if err != nil {
		return fmt.Errorf("resolve ip error")
	}

	daddr, err := net.ResolveIPAddr("ip", options.Daddr)
	if err != nil {
		return fmt.Errorf("resolve ip error")
	}

	iftype := IFTypeS11MMEGTPC
	if options.IFTypeName != "" {
		iftype, err = EnumIFTypeString(options.IFTypeName)
		if err != nil {
			return fmt.Errorf("invalid IFTypeName")
		}
	}

	conn, err := gtpv2.Dial(
		c.Ctx,
		saddr,
		daddr,
		uint8(iftype),
		uint8(options.Count),
	)
	if err != nil {
		return err
	}
	c.Conn = conn
	return nil
}

func (c *K6GTPv2) SendEchoRequest(daddr string) (uint32, error) {
	d, err := net.ResolveIPAddr("ip", daddr)
	if err != nil {
		return 0, fmt.Errorf("resolve ip error")
	}
	return c.Conn.EchoRequest(d)
}

func (c *K6GTPv2) SendEchoRequestWithCheckEchoResponce(daddr string) error {
	if _, err := c.SendEchoRequest(daddr); err != nil {
		return err
	}
	return c.RecvEchoResponce()
}

func (c *K6GTPv2) RecvEchoResponce() error {
	var err error
	buf := make([]byte, 1500)

	// if no response coming within 3 seconds, returns error without retrying.
	if err := c.Conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return 0, err
	}
	n, _, err := c.Conn.ReadFrom(buf)
	if err != nil {
		return err
	}
	if err := c.Conn.SetReadDeadline(time.Time{}); err != nil {
		return err
	}

	// decode incoming message and let it be handled by default handler funcs.
	msg, err := message.Parse(buf[:n])
	if err != nil {
		return err
	}

	if _, ok := msg.(*message.EchoResponse); !ok {
		return 0, &gtpv2.UnexpectedTypeError{Msg: msg}
	}
	return nil
}
