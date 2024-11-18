package certificateproviderpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

//go:generate enumerator -type howYouWillConfirmYourIdentity -empty -trimprefix
type howYouWillConfirmYourIdentity uint8

const (
	howYouWillConfirmYourIdentityAtPostOffice howYouWillConfirmYourIdentity = iota + 1
	howYouWillConfirmYourIdentityPostOfficeSuccessfully
	howYouWillConfirmYourIdentityOneLogin
)

type howWillYouConfirmYourIdentityData struct {
	App     appcontext.Data
	Errors  validation.List
	Form    *howWillYouConfirmYourIdentityForm
	Options howYouWillConfirmYourIdentityOptions
}

func HowWillYouConfirmYourIdentity(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *certificateproviderdata.Provided) error {
		data := &howWillYouConfirmYourIdentityData{
			App:     appData,
			Form:    &howWillYouConfirmYourIdentityForm{},
			Options: howYouWillConfirmYourIdentityValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWillYouConfirmYourIdentityForm(r, "howYouWillConfirmYourIdentity")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				switch data.Form.How {
				case howYouWillConfirmYourIdentityAtPostOffice:
					provided.Tasks.ConfirmYourIdentity = task.IdentityStatePending

					if err := certificateProviderStore.Put(r.Context(), provided); err != nil {
						return fmt.Errorf("error updating certificate provider: %w", err)
					}

					return certificateprovider.PathTaskList.Redirect(w, r, appData, provided.LpaID)

				default:
					return certificateprovider.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
				}
			}
		}

		return tmpl(w, data)
	}
}

type howWillYouConfirmYourIdentityForm struct {
	How        howYouWillConfirmYourIdentity
	errorLabel string
}

func readHowWillYouConfirmYourIdentityForm(r *http.Request, errorLabel string) *howWillYouConfirmYourIdentityForm {
	howWillYouConfirmYourIdentity, _ := ParseHowYouWillConfirmYourIdentity(page.PostFormString(r, "how"))

	return &howWillYouConfirmYourIdentityForm{
		How:        howWillYouConfirmYourIdentity,
		errorLabel: errorLabel,
	}
}

func (f *howWillYouConfirmYourIdentityForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("how", f.errorLabel, f.How,
		validation.Selected())

	return errors
}
