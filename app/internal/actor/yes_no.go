package actor

//go:generate enumerator -type YesNo -linecomment -empty
type YesNo uint8

const (
	Yes YesNo = iota + 1 // yes
	No                   // no
)
