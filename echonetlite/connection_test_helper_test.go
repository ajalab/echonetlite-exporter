package echonetlite

import (
	"net"
	"time"
)

type stubPacketConn struct {
	writeToFunc func([]byte, net.Addr) (int, error)
}

func (s *stubPacketConn) ReadFrom(_ []byte) (int, net.Addr, error) {
	return 0, nil, net.ErrClosed
}

func (s *stubPacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	if s.writeToFunc != nil {
		return s.writeToFunc(p, addr)
	}
	return len(p), nil
}

func (s *stubPacketConn) Close() error {
	return nil
}

func (s *stubPacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{}
}

func (s *stubPacketConn) SetDeadline(_ time.Time) error {
	return nil
}

func (s *stubPacketConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (s *stubPacketConn) SetWriteDeadline(_ time.Time) error {
	return nil
}
