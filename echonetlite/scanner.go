package echonetlite

import (
	"context"
	"net"
)

type Node struct {
	Host        string
	NodeProfile NodeProfile
}

type Scanner struct {
	conn *Connection
}

func NewScanner(conn *Connection) Scanner {
	return Scanner{conn: conn}
}

func (s Scanner) ScanNodes(ctx context.Context) ([]Node, error) {
	req := Frame{
		SEOJ: EOJ{0x05, 0xFF, 0x01},
		DEOJ: EOJ{0x0E, 0xF0, 0x01},
		ESV:  0x62,
		Properties: []Property{
			{EPC: 0x80, EDT: []byte{}},
			{EPC: 0xD6, EDT: []byte{}},
		},
	}
	responses, err := s.conn.broadcast(ctx, req)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	for _, res := range responses {
		var operatingStatus bool
		var instanceList []EOJ
		for _, p := range res.Frame.Properties {
			switch p.EPC {
			case 0x80:
				switch p.EDT[0] {
				case 0x30:
					operatingStatus = true
				case 0x31:
					operatingStatus = false
				}
			case 0xD6:
				edts := p.EDT
				for i := 1; i+3 <= len(edts); i += 3 {
					var instance EOJ
					copy(instance[:], edts[i:i+3])
					instanceList = append(instanceList, instance)
				}
			}
		}

		host, _, err := net.SplitHostPort(res.Addr.String())
		if err != nil {
			host = res.Addr.String()
		}
		nodes = append(nodes, Node{
			Host: host,
			NodeProfile: NodeProfile{
				OperatingStatus:       operatingStatus,
				SelfNodeInstanceListS: instanceList,
			},
		})
	}
	return nodes, nil
}
