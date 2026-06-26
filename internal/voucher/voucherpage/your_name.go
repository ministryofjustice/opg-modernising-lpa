package voucherpage

import (
	"cmp"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type yourNameData struct {
	App  appcontext.Data
	Form *yourNameForm
}

func YourName(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := yourNameData{
			App:  appData,
			Form: newYourNameForm(appData.Localizer),
		}

		data.Form.FirstNames.SetInput(cmp.Or(provided.FirstNames, lpa.Voucher.FirstNames))
		data.Form.LastName.SetInput(cmp.Or(provided.LastName, lpa.Voucher.LastName))

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			if provided.FirstNames != data.Form.FirstNames.Value || provided.LastName != data.Form.LastName.Value {
				provided.FirstNames = data.Form.FirstNames.Value
				provided.LastName = data.Form.LastName.Value

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

		return tmpl(w, data)
	}
}

type yourNameForm struct {
	forms.Form
	FirstNames *forms.String
	LastName   *forms.String
}

func newYourNameForm(l Localizer) *yourNameForm {
	return &yourNameForm{
		FirstNames: forms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: forms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
	}
}

func (f *yourNameForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.FirstNames, f.LastName)
}
