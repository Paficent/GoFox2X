package main

import (
	"log"

	"github.com/Paficent/GoFox2X/data"
	"github.com/Paficent/GoFox2X/protocol"
	"github.com/Paficent/GoFox2X/server"
	"github.com/Paficent/GoFox2X/transport"
)

const listenAddr = "0.0.0.0:9933"

func main() {
	srv := &server.Server{
		Handler: server.HandlerFunc(handle),
		OnConnect: func(c *transport.Conn) {
			log.Printf("connect    %s", c.RemoteAddr())
		},
		OnDisconnect: func(c *transport.Conn, err error) {
			log.Printf("disconnect %s (%v)", c.RemoteAddr(), err)
		},
	}

	log.Printf("SFS2X example server listening on %s", listenAddr)
	log.Fatal(srv.ListenAndServe(listenAddr))
}

func handle(c *transport.Conn, m *protocol.Message) {
	switch {
	case m.Controller == protocol.System && m.Action == protocol.ActionHandshake:
		handleHandshake(c, m)
	case m.Controller == protocol.System && m.Action == protocol.ActionLogin:
		handleLogin(c, m)
	default:
		log.Printf("unhandled  controller=%s action=%d", m.Controller, m.Action)
	}
}

func handleHandshake(c *transport.Conn, _ *protocol.Message) {
	log.Printf("handshake  %s", c.RemoteAddr())

	sessionInfo := data.MakeGFSObject().
		PutInt("ct", 1_000_000).
		PutInt("ms", 8_000_000).
		PutUtfString("tk", "0123456789abcdef0123456789abcdef")

	reply := protocol.NewMessage(protocol.System, protocol.ActionHandshake, sessionInfo)
	if err := c.Send(reply); err != nil {
		log.Printf("send handshake: %v", err)
	}
}

func handleLogin(c *transport.Conn, m *protocol.Message) {
	username, _ := m.Payload.GetUtfString("un")
	log.Printf("login      %s as %q", c.RemoteAddr(), username)

	login := data.MakeGFSObject().
		PutShort("rs", 0).
		PutUtfString("zn", "ZoneName").
		PutUtfString("un", username).
		PutShort("pi", 1).
		PutInt("id", 1).
		PutGFSObject("p", data.MakeGFSObject())
	// TODO: attach a room list under "rl" as an SFSArray of room objects.

	reply := protocol.NewMessage(protocol.System, protocol.ActionLogin, login)
	if err := c.Send(reply); err != nil {
		log.Printf("send login: %v", err)
	}
}
