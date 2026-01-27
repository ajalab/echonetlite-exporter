package echonetlite

import "fmt"

// EOJ represents an ECHONET Lite object identifier.
type EOJ [3]byte

func (e EOJ) String() string {
	return fmt.Sprintf("%02X%02X%02X", e[0], e[1], e[2])
}
