package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type canYouSignYourLpaData struct {
	App               appcontext.Data
	Errors            validation.List
	Form              *canYouSignYourLpaForm
	YesNoMaybeOptions donordata.YesNoMaybeOptions
	CanTaskList       bool
}

func CanYouSignYourLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &canYouSignYourLpaData{
			App: appData,
			Form: &canYouSignYourLpaForm{
				CanSign: provided.Donor.ThinksCanSign,
			},
			YesNoMaybeOptions: donordata.YesNoMaybeValues,
			CanTaskList:       !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readCanYouSignYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.Donor.ThinksCanSign = data.Form.CanSign

				if provided.Donor.ThinksCanSign.IsYes() {
					provided.Donor.CanSign = form.Yes
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.Donor.ThinksCanSign.IsYes() {
					return donor.PathYourPreferredLanguage.Redirect(w, r, appData, provided)
				}

				return donor.PathCheckYouCanSign.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type canYouSignYourLpaForm struct {
	CanSign donordata.YesNoMaybe
}

func readCanYouSignYourLpaForm(r *http.Request) *canYouSignYourLpaForm {
	canSign, _ := donordata.ParseYesNoMaybe(page.PostFormString(r, "can-sign"))

	return &canYouSignYourLpaForm{
		CanSign: canSign,
	}
}

func (f *canYouSignYourLpaForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("can-sign", "yesIfCanSign", f.CanSign,
		validation.Selected())

	return errors
}
