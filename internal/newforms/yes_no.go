package newforms

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

type YesNo = Enum[form.YesNo, form.YesNoOptions, *form.YesNo]

func NewYesNo(label string) *YesNo {
	return NewEnum[form.YesNo]("yesNo", label, form.YesNoValues)
}

type YesNoForm struct {
	YesNo  *YesNo
	Errors []Field
}

func NewYesNoForm(label string) *YesNoForm {
	return &YesNoForm{
		YesNo: NewYesNo(label).Selected(),
	}
}

func (f *YesNoForm) Parse(r *http.Request) bool {
	f.Errors = ParsePostForm(r, f.YesNo)

	return len(f.Errors) == 0
}
