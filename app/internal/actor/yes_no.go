package actor

import "fmt"

type YesNo string

const (
	Yes = YesNo("yes")
	No  = YesNo("no")
)

func ParseYesNo(s string) (YesNo, error) {
	switch s {
	case "yes":
		return Yes, nil
	case "no":
		return No, nil
	default:
		return YesNo(""), fmt.Errorf("invalid YesNo '%s'", s)
	}
}

func (e YesNo) IsYes() bool {
	return e == Yes
}

func (e YesNo) IsNo() bool {
	return e == No
}

func (e YesNo) String() string {
	return string(e)
}

type YesNoOptions struct {
	Yes YesNo
	No  YesNo
}

var YesNoValues = YesNoOptions{
	Yes: Yes,
	No:  No,
}
