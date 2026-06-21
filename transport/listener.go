package transport

import "net"

type Listener struct {
	ln net.Listener
}

func Listen(addr string) (*Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Listener{ln: ln}, nil
}

func (l *Listener) Accept() (*Conn, error) {
	nc, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}
	return newConn(nc), nil
}

func (l *Listener) Close() error {
	return l.ln.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.ln.Addr()
}
