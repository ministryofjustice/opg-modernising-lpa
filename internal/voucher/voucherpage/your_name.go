package voucherpage

import (
	"cmp"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type yourNameData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *yourNameForm
}

func YourName(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		firstNames := cmp.Or(provided.FirstNames, lpa.Voucher.FirstNames)
		lastName := cmp.Or(provided.LastName, lpa.Voucher.LastName)

		data := &yourNameData{
			App: appData,
			Form: &yourNameForm{
				FirstNames: firstNames,
				LastName:   lastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readYourNameForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if provided.FirstNames != data.Form.FirstNames || provided.LastName != data.Form.LastName {
					provided.FirstNames = data.Form.FirstNames
					provided.LastName = data.Form.LastName

					provided.Tasks.ConfirmYourName = task.StateInProgress

					if err := voucherStore.Put(r.Context(), provided); err != nil {
						return err
					}

					http.SetCookie(w, &http.Cookie{
						Name:     "banner",
						Value:    "1",
						MaxAge:   60,
						SameSite: http.SameSiteStrictMode,
						HttpOnly: true,
						Secure:   true,
					})
				}

				return voucher.PathConfirmYourName.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type yourNameForm struct {
	FirstNames string
	LastName   string
}

func readYourNameForm(r *http.Request) *yourNameForm {
	return &yourNameForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}
}

func (f *yourNameForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}
