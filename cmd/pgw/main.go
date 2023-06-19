package main

import (
	"context"
	"flag"
	"log"
	"net"
	"time"

	"github.com/wmnsk/go-gtp/gtpv2"
	"github.com/wmnsk/go-gtp/gtpv2/message"
)

var (
	s5c     = flag.String("s5c", "127.0.0.1", "IP for S5-C interface.")
	s5u     = flag.String("s5u", "127.0.0.1", "IP for S5-U interface.")
	s5cport = flag.String("cport", gtpv2.GTPCPort, "GTP CPort")
)

func main() {
	flag.Parse()
	log.SetPrefix("[P-GW] ")

	s5cAddr, err := net.ResolveUDPAddr("udp", *s5c+*s5cport)
	if err != nil {
		log.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start listening on the specified IP:Port.
	s5cConn := gtpv2.NewConn(s5cAddr, gtpv2.IFTypeS5S8PGWGTPC, 0)
	go func() {
		if err := s5cConn.ListenAndServe(ctx); err != nil {
			log.Println(err)
			return
		}
	}()
	log.Printf("Started serving C-Plane on %s", s5cAddr)

	// register handlers for ALL the message you expect remote endpoint to send.
	s5cConn.AddHandlers(map[uint8]gtpv2.HandlerFunc{
		message.MsgTypeCreateSessionRequest: handleCreateSessionRequest,
		message.MsgTypeDeleteSessionRequest: handleDeleteSessionRequest,
		message.MsgTypeModifyBearerRequest:  handleModifyBearerRequest,
		message.MsgTypeEchoRequest:          handleEchoRequest,
	})

	for {
		select {
		case str := <-loggerCh:
			log.Printf("%s", str)
		case err := <-errCh:
			log.Printf("Warning: %s", err)
		case <-time.After(100 * time.Second):
			var activeIMSIs []string
			for _, sess := range s5cConn.Sessions() {
				if !sess.IsActive() {
					continue
				}
				activeIMSIs = append(activeIMSIs, sess.IMSI)
			}
			if len(activeIMSIs) == 0 {
				continue
			}

			log.Println("Active Subscribers:")
			for _, imsi := range activeIMSIs {
				log.Printf("\t%s", imsi)
			}
			activeIMSIs = nil
		}
	}
}
