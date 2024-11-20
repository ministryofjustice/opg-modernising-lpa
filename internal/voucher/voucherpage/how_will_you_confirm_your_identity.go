package voucherpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

//go:generate enumerator -type howYouWillConfirmYourIdentity -empty -trimprefix
type howYouWillConfirmYourIdentity uint8

const (
	howYouWillConfirmYourIdentityAtPostOffice howYouWillConfirmYourIdentity = iota + 1
	howYouWillConfirmYourIdentityPostOfficeSuccessfully
	howYouWillConfirmYourIdentityOneLogin
)

type howWillYouConfirmYourIdentityData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.SelectForm[howYouWillConfirmYourIdentity, howYouWillConfirmYourIdentityOptions, *howYouWillConfirmYourIdentity]
}

func HowWillYouConfirmYourIdentity(tmpl template.Template, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		data := &howWillYouConfirmYourIdentityData{
			App:  appData,
			Form: form.NewEmptySelectForm[howYouWillConfirmYourIdentity](howYouWillConfirmYourIdentityValues, "howYouWillConfirmYourIdentity"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				switch data.Form.Selected {
				case howYouWillConfirmYourIdentityAtPostOffice:
					provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending

					if err := voucherStore.Put(r.Context(), provided); err != nil {
						return fmt.Errorf("error updating certificate provider: %w", err)
					}

					return voucher.PathTaskList.Redirect(w, r, appData, provided.LpaID)

				default:
					return voucher.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
				}
			}
		}

		return tmpl(w, data)
	}
}
