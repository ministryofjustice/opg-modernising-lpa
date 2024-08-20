package voucherpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type yourDeclarationData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *yourDeclarationForm
	Lpa     *lpadata.Lpa
	Voucher *voucherdata.Provided
}

func YourDeclaration(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		if !provided.SignedAt.IsZero() {
			return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourDeclarationData{
			App:     appData,
			Form:    &yourDeclarationForm{},
			Lpa:     lpa,
			Voucher: provided,
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDeclarationForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.SignedAt = now()
				provided.Tasks.SignTheDeclaration = task.StateCompleted
				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type yourDeclarationForm struct {
	Confirm bool
}

func readYourDeclarationForm(r *http.Request) *yourDeclarationForm {
	return &yourDeclarationForm{
		Confirm: page.PostFormString(r, "confirm") == "1",
	}
}

func (f *yourDeclarationForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("confirm", "youMustSelectTheBoxToVouch", f.Confirm,
		validation.Selected().CustomError())

	return errors
}
