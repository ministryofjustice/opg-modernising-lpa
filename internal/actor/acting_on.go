package actor

//go:generate enumerator -type ActingOn -linecomment -empty
type ActingOn uint8

const (
	Paper  ActingOn = iota + 1 // paper
	Online                     // online
)
