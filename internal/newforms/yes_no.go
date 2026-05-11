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
	Form
	YesNo *YesNo
}

func NewYesNoForm(label string) *YesNoForm {
	return &YesNoForm{
		YesNo: NewYesNo(label).Selected(),
	}
}

func (f *YesNoForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.YesNo)
}
