package actor

//go:generate enumerator -type YesNo -linecomment
type YesNo uint8

const (
	Unselected YesNo = iota
	Yes              // yes
	No               // no
)
