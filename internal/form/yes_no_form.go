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
	FieldName  string
}

func ReadYesNoForm(r *http.Request, errorLabel string) *YesNoForm {
	form := NewYesNoForm(YesNoUnknown)

	form.YesNo, form.Error = ParseYesNo(PostFormString(r, form.FieldName))
	form.ErrorLabel = errorLabel

	return form
}

func (f *YesNoForm) Validate() validation.List {
	var errors validation.List

	errors.Error(f.FieldName, f.ErrorLabel, f.Error,
		validation.Selected())

	return errors
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func NewYesNoForm(yesNo YesNo) *YesNoForm {
	return &YesNoForm{
		YesNo:     yesNo,
		Options:   YesNoValues,
		FieldName: FieldNames.YesNo,
	}
}
