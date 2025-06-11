package donordata

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

//go:generate go tool enumerator -type YesNoMaybe -linecomment -empty
type YesNoMaybe uint8

const (
	Yes YesNoMaybe = iota + 1
	No
	Maybe
)

type YesNoMaybeForm struct {
	errorLabel string
	Option     YesNoMaybe
}

func ReadYesNoMaybeForm(r *http.Request, errorLabel string) *YesNoMaybeForm {
	option, _ := ParseYesNoMaybe(strings.TrimSpace(r.PostFormValue("option")))

	return &YesNoMaybeForm{
		errorLabel: errorLabel,
		Option:     option,
	}
}

func (f *YesNoMaybeForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("option", f.errorLabel, f.Option,
		validation.Selected())

	return errors
}
