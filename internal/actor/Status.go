package actor

//go:generate enumerator -type Status -linecomment
type Status uint8

const (
	Active    Status = iota + 1 //active
	Suspended                   // suspended
)
