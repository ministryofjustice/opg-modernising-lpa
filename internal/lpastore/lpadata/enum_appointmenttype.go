// Code generated by "enumerator -type AppointmentType -trimprefix -linecomment"; DO NOT EDIT.

package lpadata

import (
	"fmt"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AppointmentTypeOriginal-0]
	_ = x[AppointmentTypeReplacement-1]
}

const _AppointmentType_name = "originalreplacement"

var _AppointmentType_index = [...]uint8{0, 8, 19}

func (i AppointmentType) String() string {
	if i >= AppointmentType(len(_AppointmentType_index)-1) {
		return "AppointmentType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AppointmentType_name[_AppointmentType_index[i]:_AppointmentType_index[i+1]]
}

func (i AppointmentType) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *AppointmentType) UnmarshalText(text []byte) error {
	val, err := ParseAppointmentType(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i AppointmentType) IsOriginal() bool {
	return i == AppointmentTypeOriginal
}

func (i AppointmentType) IsReplacement() bool {
	return i == AppointmentTypeReplacement
}

func ParseAppointmentType(s string) (AppointmentType, error) {
	switch s {
	case "original":
		return AppointmentTypeOriginal, nil
	case "replacement":
		return AppointmentTypeReplacement, nil
	default:
		return AppointmentType(0), fmt.Errorf("invalid AppointmentType '%s'", s)
	}
}

type AppointmentTypeOptions struct {
	Original    AppointmentType
	Replacement AppointmentType
}

var AppointmentTypeValues = AppointmentTypeOptions{
	Original:    AppointmentTypeOriginal,
	Replacement: AppointmentTypeReplacement,
}
