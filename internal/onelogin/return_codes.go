package onelogin

//go:generate enumerator -type ReturnCode --trimprefix
type ReturnCode uint8

const (
	ReturnCodeUnknown ReturnCode = iota
	ReturnCodeA
	ReturnCodeD
	ReturnCodeN
	ReturnCodeP
	ReturnCodeT
	ReturnCodeV
	ReturnCodeX
	ReturnCodeZ
)
