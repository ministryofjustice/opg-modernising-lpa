// Code generated by "enumerator -type AttorneyStatus -trimprefix -linecomment"; DO NOT EDIT.

package lpadata

import (
	"fmt"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AttorneyStatusActive-0]
	_ = x[AttorneyStatusInactive-1]
	_ = x[AttorneyStatusRemoved-2]
}

const _AttorneyStatus_name = "activeinactiveremoved"

var _AttorneyStatus_index = [...]uint8{0, 6, 14, 21}

func (i AttorneyStatus) String() string {
	if i >= AttorneyStatus(len(_AttorneyStatus_index)-1) {
		return "AttorneyStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AttorneyStatus_name[_AttorneyStatus_index[i]:_AttorneyStatus_index[i+1]]
}

func (i AttorneyStatus) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *AttorneyStatus) UnmarshalText(text []byte) error {
	val, err := ParseAttorneyStatus(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i AttorneyStatus) IsActive() bool {
	return i == AttorneyStatusActive
}

func (i AttorneyStatus) IsInactive() bool {
	return i == AttorneyStatusInactive
}

func (i AttorneyStatus) IsRemoved() bool {
	return i == AttorneyStatusRemoved
}

func ParseAttorneyStatus(s string) (AttorneyStatus, error) {
	switch s {
	case "active":
		return AttorneyStatusActive, nil
	case "inactive":
		return AttorneyStatusInactive, nil
	case "removed":
		return AttorneyStatusRemoved, nil
	default:
		return AttorneyStatus(0), fmt.Errorf("invalid AttorneyStatus '%s'", s)
	}
}

type AttorneyStatusOptions struct {
	Active   AttorneyStatus
	Inactive AttorneyStatus
	Removed  AttorneyStatus
}

var AttorneyStatusValues = AttorneyStatusOptions{
	Active:   AttorneyStatusActive,
	Inactive: AttorneyStatusInactive,
	Removed:  AttorneyStatusRemoved,
}
