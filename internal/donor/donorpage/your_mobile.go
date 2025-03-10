package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourMobileData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *yourMobileForm
	CanTaskList bool
}

func YourMobile(tmpl template.Template, donorStore DonorStore, shareCodeSender ShareCodeSender) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourMobileData{
			App: appData,
			Form: &yourMobileForm{
				Mobile: provided.Donor.Mobile,
			},
			CanTaskList: !provided.Type.Empty(),
		}

		if r.Method == http.MethodPost {
			data.Form = readYourMobileForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirect := donor.PathReceivingUpdatesAboutYourLpa
				provided.Donor.Mobile = data.Form.Mobile

				if shareCodeSender != nil {
					redirect = donor.PathWeHaveContactedVoucher

					if err := shareCodeSender.SendVoucherAccessCode(r.Context(), provided, appData); err != nil {
						return fmt.Errorf("error sending voucher access code: %w", err)
					}
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return redirect.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourMobileForm struct {
	Mobile string
}

func readYourMobileForm(r *http.Request) *yourMobileForm {
	return &yourMobileForm{Mobile: page.PostFormString(r, "mobile")}
}

func (f *yourMobileForm) Validate() validation.List {
	var errors validation.List

	errors.String("mobile", "mobile", f.Mobile,
		validation.Mobile())

	return errors
}
