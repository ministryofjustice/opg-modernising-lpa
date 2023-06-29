package form

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type HappyForm struct {
	Happy actor.YesNo
	Error error
}

func ReadHappyForm(r *http.Request) *HappyForm {
	happy, err := actor.ParseYesNo(page.PostFormString(r, "happy"))

	return &HappyForm{
		Happy: happy,
		Error: err,
	}
}

func (f *HappyForm) Validate(label string) validation.List {
	var errors validation.List

	errors.Error("happy", label, f.Error,
		validation.Selected())

	return errors
}
