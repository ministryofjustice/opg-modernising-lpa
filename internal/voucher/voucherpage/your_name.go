package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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

		firstNames := provided.FirstNames
		if firstNames == "" {
			firstNames = lpa.Voucher.FirstNames
		}

		lastName := provided.LastName
		if lastName == "" {
			lastName = lpa.Voucher.LastName
		}

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
				provided.FirstNames = data.Form.FirstNames
				provided.LastName = data.Form.LastName

				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return err
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
