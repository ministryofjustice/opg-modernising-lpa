package donorpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterVoucherData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterVoucherForm
}

func EnterVoucher(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &enterVoucherData{
			App: appData,
			Form: &enterVoucherForm{
				FirstNames: donor.Voucher.FirstNames,
				LastName:   donor.Voucher.LastName,
				Email:      donor.Voucher.Email,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterVoucherForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if donor.Voucher.FirstNames != data.Form.FirstNames || donor.Voucher.LastName != data.Form.LastName {
					donor.Voucher.FirstNames = data.Form.FirstNames
					donor.Voucher.LastName = data.Form.LastName
					donor.Voucher.Allowed = len(donor.Voucher.Matches(donor)) == 0 && !strings.EqualFold(donor.Voucher.LastName, donor.Donor.LastName)
				}

				donor.Voucher.Email = data.Form.Email

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if !donor.Voucher.Allowed {
					return page.Paths.ConfirmPersonAllowedToVouch.Redirect(w, r, appData, donor)
				}

				return page.Paths.CheckYourDetails.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type enterVoucherForm struct {
	FirstNames string
	LastName   string
	Email      string
}

func readEnterVoucherForm(r *http.Request) *enterVoucherForm {
	return &enterVoucherForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
		Email:      page.PostFormString(r, "email"),
	}
}

func (f *enterVoucherForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("email", "email", f.Email,
		validation.Empty(),
		validation.Email())

	return errors
}
