package echonetlite

import (
	"context"
	"fmt"
	"log/slog"
	"net"
)

type Node struct {
	Host        string
	NodeProfile NodeProfile
}

type NodeProfileClient struct {
	conn *Connection
}

func NewNodeProfileClient(conn *Connection) NodeProfileClient {
	return NodeProfileClient{conn: conn}
}

func (c NodeProfileClient) Scan(ctx context.Context) ([]Node, error) {
	req := newGetRequest(EOJ{0x0E, 0xF0, 0x01}, []uint8{0x80, 0xD6})
	responses, err := c.conn.Multicast(ctx, req)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	for _, res := range responses {
		profile, err := decodeNodeProfile(req, res.Frame)
		if err != nil {
			slog.Warn("ignoring invalid node profile response", "addr", res.Addr.String(), "err", err)
			continue
		}
		host, _, err := net.SplitHostPort(res.Addr.String())
		if err != nil {
			host = res.Addr.String()
		}
		nodes = append(nodes, Node{Host: host, NodeProfile: profile})
	}
	return nodes, nil
}

func decodeNodeProfile(req Frame, res *Frame) (NodeProfile, error) {
	values, err := getResponseValues(req, res)
	if err != nil {
		return NodeProfile{}, err
	}
	status, err := decodeByte(values[0x80], 0x30, 0x31)
	if err != nil {
		return NodeProfile{}, fmt.Errorf("operating status (EPC 0x80): %w", err)
	}
	instances := values[0xD6]
	if len(instances) == 0 || len(instances) != 1+3*int(instances[0]) {
		return NodeProfile{}, fmt.Errorf("invalid instance list length (EPC 0xD6)")
	}
	profile := NodeProfile{OperatingStatus: status == 0x30}
	for i := 1; i < len(instances); i += 3 {
		profile.SelfNodeInstanceListS = append(profile.SelfNodeInstanceListS, EOJ(instances[i:i+3]))
	}
	return profile, nil
}
