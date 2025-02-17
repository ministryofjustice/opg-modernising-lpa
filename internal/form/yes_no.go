package form

//go:generate go tool enumerator -type YesNo -linecomment -trimprefix
type YesNo uint8

const (
	YesNoUnknown YesNo = iota
	Yes                // yes
	No                 // no
)
