package echonetlite

type Device struct {
	host string
	eoj  EOJ
}

func NewDevice(host string, eoj EOJ) Device {
	return Device{
		host: host,
		eoj:  eoj,
	}
}

func (d Device) Host() string {
	return d.host
}

func (d Device) EOJ() EOJ {
	return d.eoj
}
