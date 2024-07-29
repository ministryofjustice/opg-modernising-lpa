package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type canYouSignYourLpaData struct {
	App               page.AppData
	Errors            validation.List
	Form              *canYouSignYourLpaForm
	YesNoMaybeOptions donordata.YesNoMaybeOptions
	CanTaskList       bool
}

func CanYouSignYourLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &canYouSignYourLpaData{
			App: appData,
			Form: &canYouSignYourLpaForm{
				CanSign: donor.Donor.ThinksCanSign,
			},
			YesNoMaybeOptions: donordata.YesNoMaybeValues,
			CanTaskList:       !donor.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Donor.ThinksCanSign = data.Form.CanSign

				if donor.Donor.ThinksCanSign.IsYes() {
					donor.Donor.CanSign = form.Yes
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.Donor.ThinksCanSign.IsYes() {
					return page.Paths.YourPreferredLanguage.Redirect(w, r, appData, donor)
				}

				return page.Paths.CheckYouCanSign.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type canYouSignYourLpaForm struct {
	CanSign      actor.YesNoMaybe
	CanSignError error
}

func readYourDetailsForm(r *http.Request) *canYouSignYourLpaForm {
	canSign, canSignError := donordata.ParseYesNoMaybe(page.PostFormString(r, "can-sign"))

	return &canYouSignYourLpaForm{
		CanSign:      canSign,
		CanSignError: canSignError,
	}
}

func (f *canYouSignYourLpaForm) Validate() validation.List {
	var errors validation.List

	errors.Error("can-sign", "yesIfCanSign", f.CanSignError,
		validation.Selected())

	return errors
}
