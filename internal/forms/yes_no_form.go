package forms

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

var (
	YesNoUnknown = form.YesNoUnknown
	Yes          = form.Yes
	No           = form.No
)

type YesNoForm struct {
	Form
	YesNo *Enum[form.YesNo, form.YesNoOptions, *form.YesNo]
}

func NewYesNoForm(label string) *YesNoForm {
	return &YesNoForm{
		YesNo: NewEnum[form.YesNo]("yesNo", label, form.YesNoValues).Selected(),
	}
}

func (f *YesNoForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.YesNo)
}
