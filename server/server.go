/*
 * Minimal usage:
 *   srv := &server.Server{
 *   	Handler: server.HandlerFunc(func(c *transport.Conn, m *protocol.Message) {
 * 			// inspect m.Controler / m.Action
 * 			// reply with c.Send(...)
 * 		})
 *   }
 *   log.Fatal(srv.ListenAndServer("0.0.0.0:9933"))
 */

package server

import (
	"github.com/Paficent/GoFox2X/protocol"
	"github.com/Paficent/GoFox2X/transport"
)

type Handler interface {
	HandleMessage(conn *transport.Conn, msg *protocol.Message)
}

type HandlerFunc func(conn *transport.Conn, msg *protocol.Message)

func (f HandlerFunc) HandleMessage(conn *transport.Conn, msg *protocol.Message) {
	f(conn, msg)
}

type Server struct {
	Handler Handler

	OnConnect func(conn *transport.Conn)

	OnDisconnect func(conn *transport.Conn, err error)
}

// binds addr and serves connections until the listener errors
func (s *Server) ListenAndServe(addr string) error {
	ln, err := transport.Listen(addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	return s.Serve(ln)
}

func (s *Server) Serve(ln *transport.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.serveConn(conn)
	}
}

// runs a single connection's lifecycle
func (s *Server) serveConn(conn *transport.Conn) {
	if s.OnConnect != nil {
		s.OnConnect(conn)
	}

	var loopErr error
	for {
		msg, err := conn.Receive()
		if err != nil {
			loopErr = err
			break
		}
		if s.Handler != nil {
			s.Handler.HandleMessage(conn, msg)
		}
	}

	_ = conn.Close()
	if s.OnDisconnect != nil {
		s.OnDisconnect(conn, loopErr)
	}
}
