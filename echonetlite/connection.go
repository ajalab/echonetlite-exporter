package echonetlite

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"syscall"

	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

type Response struct {
	Frame *Frame
	Addr  net.Addr
}

type Connection struct {
	conn          net.PacketConn
	connMulticast *ipv4.PacketConn
	pending       sync.Map
	lastTID       uint32
}

var addrMulticast = &net.UDPAddr{IP: net.ParseIP("224.0.23.0"), Port: 3610}

func NewConnection(multicastInterface string) (*Connection, error) {
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

	connMulticast := ipv4.NewPacketConn(conn.(*net.UDPConn))
	if multicastInterface != "" {
		iface, err := net.InterfaceByName("eth0")
		if err != nil {
			return nil, fmt.Errorf("failed to get interface: %v", err)
		}
		if err := connMulticast.SetMulticastInterface(iface); err != nil {
			return nil, fmt.Errorf("failed to set multicast interface (iface: %v): %v", iface, err)
		}
	}
	if err := connMulticast.SetMulticastTTL(1); err != nil {
		return nil, fmt.Errorf("failed to set multicast TTL: %v", err)
	}

	m := &Connection{conn: conn, connMulticast: connMulticast}
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
			slog.Warn("ignoring broken frame", "err", err, "addr", addr.String())
			continue
		}

		if frame.ESV >= 0x60 && frame.ESV <= 0x6E {
			slog.Info(
				"ignoring request frame",
				"esv", frame.ESV,
				"seoj", frame.SEOJ.String(),
				"deoj", frame.DEOJ.String(),
				"addr", addr.String(),
			)
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

func (c *Connection) multicast(ctx context.Context, req Frame) ([]Response, error) {
	req.TID = uint16(atomic.AddUint32(&c.lastTID, 1) & 0xFFFF)

	resCh := make(chan Response, 100)
	c.pending.Store(req.TID, resCh)
	defer c.pending.Delete(req.TID)

	if _, err := c.connMulticast.WriteTo(req.Serialize(), nil, addrMulticast); err != nil {
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
