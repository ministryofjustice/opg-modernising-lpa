package supporter

//go:generate enumerator -type Permission -linecomment -empty
type Permission uint8

const (
	Admin Permission = iota + 1 // admin
)
