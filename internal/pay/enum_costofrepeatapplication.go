// Code generated by "enumerator -type CostOfRepeatApplication -empty -trimprefix"; DO NOT EDIT.

package pay

import (
	"fmt"
	"strconv"
)

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[CostOfRepeatApplicationNoFee-1]
	_ = x[CostOfRepeatApplicationHalfFee-2]
}

const _CostOfRepeatApplication_name = "NoFeeHalfFee"

var _CostOfRepeatApplication_index = [...]uint8{0, 5, 12}

func (i CostOfRepeatApplication) String() string {
	if i == 0 {
		return ""
	}
	i -= 1
	if i >= CostOfRepeatApplication(len(_CostOfRepeatApplication_index)-1) {
		return "CostOfRepeatApplication(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _CostOfRepeatApplication_name[_CostOfRepeatApplication_index[i]:_CostOfRepeatApplication_index[i+1]]
}

func (i CostOfRepeatApplication) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

func (i *CostOfRepeatApplication) UnmarshalText(text []byte) error {
	val, err := ParseCostOfRepeatApplication(string(text))
	if err != nil {
		return err
	}

	*i = val
	return nil
}

func (i CostOfRepeatApplication) IsNoFee() bool {
	return i == CostOfRepeatApplicationNoFee
}

func (i CostOfRepeatApplication) IsHalfFee() bool {
	return i == CostOfRepeatApplicationHalfFee
}

func ParseCostOfRepeatApplication(s string) (CostOfRepeatApplication, error) {
	switch s {
	case "":
		return CostOfRepeatApplication(0), nil
	case "NoFee":
		return CostOfRepeatApplicationNoFee, nil
	case "HalfFee":
		return CostOfRepeatApplicationHalfFee, nil
	default:
		return CostOfRepeatApplication(0), fmt.Errorf("invalid CostOfRepeatApplication '%s'", s)
	}
}

type CostOfRepeatApplicationOptions struct {
	NoFee   CostOfRepeatApplication
	HalfFee CostOfRepeatApplication
}

var CostOfRepeatApplicationValues = CostOfRepeatApplicationOptions{
	NoFee:   CostOfRepeatApplicationNoFee,
	HalfFee: CostOfRepeatApplicationHalfFee,
}

func (i CostOfRepeatApplication) Empty() bool {
	return i == CostOfRepeatApplication(0)
}