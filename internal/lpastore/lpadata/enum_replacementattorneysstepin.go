// Code generated by "enumerator -type ReplacementAttorneysStepIn -linecomment -trimprefix -empty"; DO NOT EDIT.

package lpadata

import (
	"fmt"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ReplacementAttorneysStepInWhenAllCanNoLongerAct-1]
	_ = x[ReplacementAttorneysStepInWhenOneCanNoLongerAct-2]
	_ = x[ReplacementAttorneysStepInAnotherWay-3]
}

const _ReplacementAttorneysStepIn_name = "all-can-no-longer-actone-can-no-longer-actanother-way"

var _ReplacementAttorneysStepIn_index = [...]uint8{0, 21, 42, 53}

func (i ReplacementAttorneysStepIn) String() string {
	if i == 0 {
		return ""
	}
	i -= 1
	if i >= ReplacementAttorneysStepIn(len(_ReplacementAttorneysStepIn_index)-1) {
		return "ReplacementAttorneysStepIn(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ReplacementAttorneysStepIn_name[_ReplacementAttorneysStepIn_index[i]:_ReplacementAttorneysStepIn_index[i+1]]
}

func (i ReplacementAttorneysStepIn) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *ReplacementAttorneysStepIn) UnmarshalText(text []byte) error {
	val, err := ParseReplacementAttorneysStepIn(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i ReplacementAttorneysStepIn) IsWhenAllCanNoLongerAct() bool {
	return i == ReplacementAttorneysStepInWhenAllCanNoLongerAct
}

func (i ReplacementAttorneysStepIn) IsWhenOneCanNoLongerAct() bool {
	return i == ReplacementAttorneysStepInWhenOneCanNoLongerAct
}

func (i ReplacementAttorneysStepIn) IsAnotherWay() bool {
	return i == ReplacementAttorneysStepInAnotherWay
}

func ParseReplacementAttorneysStepIn(s string) (ReplacementAttorneysStepIn, error) {
	switch s {
	case "all-can-no-longer-act":
		return ReplacementAttorneysStepInWhenAllCanNoLongerAct, nil
	case "one-can-no-longer-act":
		return ReplacementAttorneysStepInWhenOneCanNoLongerAct, nil
	case "another-way":
		return ReplacementAttorneysStepInAnotherWay, nil
	default:
		return ReplacementAttorneysStepIn(0), fmt.Errorf("invalid ReplacementAttorneysStepIn '%s'", s)
	}
}

type ReplacementAttorneysStepInOptions struct {
	WhenAllCanNoLongerAct ReplacementAttorneysStepIn
	WhenOneCanNoLongerAct ReplacementAttorneysStepIn
	AnotherWay            ReplacementAttorneysStepIn
}

var ReplacementAttorneysStepInValues = ReplacementAttorneysStepInOptions{
	WhenAllCanNoLongerAct: ReplacementAttorneysStepInWhenAllCanNoLongerAct,
	WhenOneCanNoLongerAct: ReplacementAttorneysStepInWhenOneCanNoLongerAct,
	AnotherWay:            ReplacementAttorneysStepInAnotherWay,
}

func (i ReplacementAttorneysStepIn) Empty() bool {
	return i == ReplacementAttorneysStepIn(0)
}