package donorpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterVoucherData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterVoucherForm
}

func EnterVoucher(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &enterVoucherData{
			App: appData,
			Form: &enterVoucherForm{
				FirstNames: provided.Voucher.FirstNames,
				LastName:   provided.Voucher.LastName,
				Email:      provided.Voucher.Email,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterVoucherForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if provided.Voucher.UID.IsZero() {
					provided.Voucher.UID = newUID()
				}

				if provided.Voucher.FirstNames != data.Form.FirstNames || provided.Voucher.LastName != data.Form.LastName {
					provided.Voucher.FirstNames = data.Form.FirstNames
					provided.Voucher.LastName = data.Form.LastName
					provided.Voucher.Allowed = len(provided.Voucher.Matches(provided)) == 0 && !strings.EqualFold(provided.Voucher.LastName, provided.Donor.LastName)
				}

				provided.Voucher.Email = data.Form.Email

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if !provided.Voucher.Allowed {
					return donor.PathConfirmPersonAllowedToVouch.Redirect(w, r, appData, provided)
				}

				return donor.PathCheckYourDetails.Redirect(w, r, appData, provided)
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
