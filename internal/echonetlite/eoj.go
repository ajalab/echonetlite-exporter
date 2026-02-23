package echonetlite

import "fmt"

const (
	ClassPowerDistributionBoardMetering = 0x0287
	ClassPVPowerGeneration              = 0x0279
	ClassStorageBattery                 = 0x027D
	ClassMultipleInputPCS               = 0x02A5
)

// EOJ represents an ECHONET Lite object identifier.
type EOJ [3]byte

func (e EOJ) String() string {
	return fmt.Sprintf("%02X%02X%02X", e[0], e[1], e[2])
}

func (e EOJ) Class() uint16 {
	return uint16(e[0])<<8 | uint16(e[1])
}
