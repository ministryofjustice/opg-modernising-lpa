package form

//go:generate enumerator -type YesNo -linecomment -empty
type YesNo uint8

const (
	YesNoUnknown YesNo = iota
	Yes                // yes
	No                 // no
)
