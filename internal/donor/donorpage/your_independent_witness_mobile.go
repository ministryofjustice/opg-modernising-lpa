package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourIndependentWitnessMobileData struct {
	App                 appcontext.Data
	Errors              validation.List
	AuthorisedSignatory donordata.AuthorisedSignatory
	Form                *yourIndependentWitnessMobileForm
}

func YourIndependentWitnessMobile(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourIndependentWitnessMobileData{
			App:                 appData,
			AuthorisedSignatory: provided.AuthorisedSignatory,
			Form: &yourIndependentWitnessMobileForm{
				HasNonUKMobile: provided.IndependentWitness.HasNonUKMobile,
			},
		}

		if provided.IndependentWitness.HasNonUKMobile {
			data.Form.NonUKMobile = provided.IndependentWitness.Mobile
		} else {
			data.Form.Mobile = provided.IndependentWitness.Mobile
		}

		if r.Method == http.MethodPost {
			data.Form = readYourIndependentWitnessMobileForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.IndependentWitness.HasNonUKMobile = data.Form.HasNonUKMobile
				if data.Form.HasNonUKMobile {
					provided.IndependentWitness.Mobile = data.Form.NonUKMobile
				} else {
					provided.IndependentWitness.Mobile = data.Form.Mobile
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathYourIndependentWitnessAddress.Redirect(w, r, appData, provided)
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
