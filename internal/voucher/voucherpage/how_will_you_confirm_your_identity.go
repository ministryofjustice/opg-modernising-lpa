package voucherpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

//go:generate go tool enumerator -type howYouWillConfirmYourIdentity -empty -trimprefix
type howYouWillConfirmYourIdentity uint8

const (
	howYouWillConfirmYourIdentityAtPostOffice howYouWillConfirmYourIdentity = iota + 1
	howYouWillConfirmYourIdentityPostOfficeSuccessfully
	howYouWillConfirmYourIdentityOneLogin
)

type howWillYouConfirmYourIdentityData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *newforms.EnumForm[howYouWillConfirmYourIdentity, howYouWillConfirmYourIdentityOptions, *howYouWillConfirmYourIdentity]
}

func HowWillYouConfirmYourIdentity(tmpl template.Template, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		data := &howWillYouConfirmYourIdentityData{
			App:  appData,
			Form: newforms.NewEnumForm[howYouWillConfirmYourIdentity](appData.Localizer.T("howYouWillConfirmYourIdentity"), howYouWillConfirmYourIdentityValues),
		}

		if r.Method == http.MethodPost {
			if ok := data.Form.Parse(r); ok {
				switch data.Form.Enum.Value {
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
