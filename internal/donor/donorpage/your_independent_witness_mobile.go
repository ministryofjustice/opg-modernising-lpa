package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourIndependentWitnessMobileData struct {
	App                 page.AppData
	Errors              validation.List
	AuthorisedSignatory actor.AuthorisedSignatory
	Form                *yourIndependentWitnessMobileForm
}

func YourIndependentWitnessMobile(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourIndependentWitnessMobileData{
			App:                 appData,
			AuthorisedSignatory: donor.AuthorisedSignatory,
			Form: &yourIndependentWitnessMobileForm{
				HasNonUKMobile: donor.IndependentWitness.HasNonUKMobile,
			},
		}

		if donor.IndependentWitness.HasNonUKMobile {
			data.Form.NonUKMobile = donor.IndependentWitness.Mobile
		} else {
			data.Form.Mobile = donor.IndependentWitness.Mobile
		}

		if r.Method == http.MethodPost {
			data.Form = readYourIndependentWitnessMobileForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.IndependentWitness.HasNonUKMobile = data.Form.HasNonUKMobile
				if data.Form.HasNonUKMobile {
					donor.IndependentWitness.Mobile = data.Form.NonUKMobile
				} else {
					donor.IndependentWitness.Mobile = data.Form.Mobile
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.YourIndependentWitnessAddress.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type yourIndependentWitnessMobileForm struct {
	Mobile         string
	HasNonUKMobile bool
	NonUKMobile    string
}

func readYourIndependentWitnessMobileForm(r *http.Request) *yourIndependentWitnessMobileForm {
	return &yourIndependentWitnessMobileForm{
		Mobile:         page.PostFormString(r, "mobile"),
		HasNonUKMobile: page.PostFormString(r, "has-non-uk-mobile") == "1",
		NonUKMobile:    page.PostFormString(r, "non-uk-mobile"),
	}
}

func (d *yourIndependentWitnessMobileForm) Validate() validation.List {
	var errors validation.List

	if d.HasNonUKMobile {
		errors.String("non-uk-mobile", "aMobilePhoneNumber", d.NonUKMobile,
			validation.Empty(),
			validation.NonUKMobile().ErrorLabel("enterAMobileNumberInTheCorrectFormat"))
	} else {
		errors.String("mobile", "aUKMobileNumber", d.Mobile,
			validation.Empty(),
			validation.Mobile().ErrorLabel("enterAMobileNumberInTheCorrectFormat"))
	}

	return errors
}
