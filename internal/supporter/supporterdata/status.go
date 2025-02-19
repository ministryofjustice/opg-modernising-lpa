package supporterdata

//go:generate go tool enumerator -type Status -linecomment -trimprefix
type Status uint8

const (
	StatusActive    Status = iota // active
	StatusSuspended               // suspended
)
