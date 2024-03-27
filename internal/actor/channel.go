package actor

//go:generate enumerator -type Channel -linecomment -empty
type Channel uint8

const (
	Paper  Channel = iota + 1 // paper
	Online                    // online
)
