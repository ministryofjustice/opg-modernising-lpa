package form

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type YesNoForm struct {
	YesNo      YesNo
	Error      error
	ErrorLabel string
	Options    YesNoOptions
}

func ReadYesNoForm(r *http.Request, errorLabel string) *YesNoForm {
	yesNo, err := ParseYesNo(PostFormString(r, "yes-no"))

	return &YesNoForm{
		YesNo:      yesNo,
		Error:      err,
		ErrorLabel: errorLabel,
		Options:    YesNoValues,
	}
}

func (f *YesNoForm) Validate() validation.List {
	var errors validation.List

	errors.Error("yes-no", f.ErrorLabel, f.Error,
		validation.Selected())

	return errors
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func NewYesNoForm(yesNo YesNo) *YesNoForm {
	return &YesNoForm{
		YesNo:   yesNo,
		Options: YesNoValues,
	}
}
