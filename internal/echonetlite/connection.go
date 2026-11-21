package echonetlite

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

type Connection struct {
	conn         net.PacketConn
	pending      sync.Map
	lastTID      uint32
	responseWait time.Duration
	requestGates sync.Map
}

type pendingRequest struct {
	ctx       context.Context
	sourceIP  net.IP
	responses chan FrameAddr
}

var addrMulticast = &net.UDPAddr{IP: net.ParseIP("224.0.23.0"), Port: 3610}

func NewConnection(multicastInterface string, responseWait time.Duration) (*Connection, error) {
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

	// Multicast options apply to the underlying socket used by conn.WriteTo.
	ipv4Conn := ipv4.NewPacketConn(conn)
	if multicastInterface != "" {
		iface, err := net.InterfaceByName(multicastInterface)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to get interface: %v", err)
		}
		if err := ipv4Conn.SetMulticastInterface(iface); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to set multicast interface (iface: %v): %v", iface, err)
		}
	}
	if err := ipv4Conn.SetMulticastTTL(1); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to set multicast TTL: %v", err)
	}

	m := &Connection{conn: conn, responseWait: responseWait}
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

		c.dispatchResponse(frame, addr)
	}
}

func (c *Connection) dispatchResponse(frame *Frame, addr net.Addr) {
	value, ok := c.pending.Load(frame.TID)
	if !ok {
		return
	}
	pending := value.(*pendingRequest)
	response := FrameAddr{Frame: frame, Addr: addr}
	if pending.sourceIP != nil {
		peer, ok := addr.(*net.UDPAddr)
		// The ECHONET Lite UDP source port is unspecified.
		if !ok || !peer.IP.Equal(pending.sourceIP) {
			return
		}
		// A delayed duplicate must never block reception for other requests.
		select {
		case pending.responses <- response:
		default:
		}
		return
	}
	select {
	case pending.responses <- response:
	case <-pending.ctx.Done():
	}
}

func (c *Connection) registerRequest(pending *pendingRequest) (uint16, error) {
	for range 1 << 16 {
		tid := uint16(atomic.AddUint32(&c.lastTID, 1))
		if _, used := c.pending.LoadOrStore(tid, pending); !used {
			return tid, nil
		}
	}
	return 0, fmt.Errorf("no available transaction ID")
}

func (c *Connection) unicast(ctx context.Context, host string, req Frame, minimumWait time.Duration) (*Frame, error) {
	addr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(host, "3610"))
	if err != nil {
		return nil, fmt.Errorf("resolve addr error: %w", err)
	}

	// AIF continuous-request timing applies to the node, across its EOJs.
	value, _ := c.requestGates.LoadOrStore(addr.String(), make(chan struct{}, 1))
	gate := value.(chan struct{})
	select {
	case gate <- struct{}{}:
		defer func() { <-gate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	pending := &pendingRequest{sourceIP: addr.IP, responses: make(chan FrameAddr, 1)}
	req.TID, err = c.registerRequest(pending)
	if err != nil {
		return nil, err
	}
	defer c.pending.Delete(req.TID)

	if _, err := c.conn.WriteTo(req.Serialize(), addr); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, max(c.responseWait, minimumWait))
	defer cancel()
	select {
	case res := <-pending.responses:
		return res.Frame, nil
	case <-waitCtx.Done():
		return nil, waitCtx.Err()
	}
}

func (c *Connection) Multicast(ctx context.Context, req Frame) ([]FrameAddr, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	pending := &pendingRequest{ctx: ctx, responses: make(chan FrameAddr, 100)}
	var err error
	req.TID, err = c.registerRequest(pending)
	if err != nil {
		return nil, err
	}
	defer c.pending.Delete(req.TID)

	if _, err := c.conn.WriteTo(req.Serialize(), addrMulticast); err != nil {
		return nil, fmt.Errorf("write error: %w", err)
	}

	var results []FrameAddr

	for {
		select {
		case res := <-pending.responses:
			results = append(results, res)
		case <-ctx.Done():
			return results, nil
		}
	}
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
