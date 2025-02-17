package lpadata

//go:generate go tool enumerator -type AppointmentType -trimprefix -linecomment
type AppointmentType uint8

const (
	AppointmentTypeOriginal    AppointmentType = iota // original
	AppointmentTypeReplacement                        // replacement
)
