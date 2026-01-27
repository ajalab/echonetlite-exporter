package echonetlite

type NodeProfile struct {
	// Indicates the node operating status.
	// EPC: 0x80
	OperatingStatus bool

	// Self-node instance list
	// EPC: 0xD6
	SelfNodeInstanceListS []EOJ
}
