package supporterdata

//go:generate enumerator -type Permission -linecomment -trimprefix
type Permission uint8

const (
	PermissionNone  Permission = iota // none
	PermissionAdmin                   // admin
)
