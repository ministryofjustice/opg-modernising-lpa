package actor

//go:generate enumerator -type Permission -linecomment
type Permission uint8

const (
	None  Permission = iota //none
	Admin                   // admin
)
