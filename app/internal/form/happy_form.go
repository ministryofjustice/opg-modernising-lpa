package form

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type HappyForm struct {
	Happy string
}

func ReadHappyForm(r *http.Request) *HappyForm {
	return &HappyForm{
		Happy: page.PostFormString(r, "happy"),
	}
}

func (f *HappyForm) Validate(label string) validation.List {
	var errors validation.List

	errors.String("happy", label, f.Happy,
		validation.Select("yes", "no"))

	return errors
}
