package echonetlite

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"syscall"

	"golang.org/x/sys/unix"
)

type Response struct {
	Frame *Frame
	Addr  net.Addr
}

type Connection struct {
	conn    net.PacketConn
	pending sync.Map
	lastTID uint32
}

func NewConnection() (*Connection, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var err error
			c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					return
				}
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
			return err
		},
	}
	conn, err := lc.ListenPacket(context.Background(), "udp4", "0.0.0.0:3610")
	if err != nil {
		return nil, err
	}

	m := &Connection{conn: conn}
	go m.receiveLoop()
	return m, nil
}

func (c *Connection) receiveLoop() {
	buf := make([]byte, 65535)
	for {
		n, addr, err := c.conn.ReadFrom(buf)
		if err != nil {
			return
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])
		frame, err := Deserialize(packet)
		if err != nil {
			continue
		}

		if ch, ok := c.pending.Load(frame.TID); ok {
			ch.(chan Response) <- Response{Frame: frame, Addr: addr}
		}
	}
}

func (c *Connection) unicast(ctx context.Context, host string, req Frame) (*Frame, error) {
	req.TID = uint16(atomic.AddUint32(&c.lastTID, 1) & 0xFFFF)

	resCh := make(chan Response, 1)
	c.pending.Store(req.TID, resCh)
	defer c.pending.Delete(req.TID)

	addr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(host, "3610"))
	if err != nil {
		return nil, fmt.Errorf("resolve addr error: %w", err)
	}

	if _, err := c.conn.WriteTo(req.Serialize(), addr); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	select {
	case res := <-resCh:
		return res.Frame, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *Connection) broadcast(ctx context.Context, req Frame) ([]Response, error) {
	dest := "224.0.23.0:3610"
	req.TID = uint16(atomic.AddUint32(&c.lastTID, 1) & 0xFFFF)

	resCh := make(chan Response, 100)
	c.pending.Store(req.TID, resCh)
	defer c.pending.Delete(req.TID)

	addr, _ := net.ResolveUDPAddr("udp4", dest)
	if _, err := c.conn.WriteTo(req.Serialize(), addr); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	var results []Response

	for {
		select {
		case res := <-resCh:
			results = append(results, Response{
				Frame: res.Frame,
				Addr:  res.Addr,
			})
		case <-ctx.Done():
			return results, nil
		}
	}
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
