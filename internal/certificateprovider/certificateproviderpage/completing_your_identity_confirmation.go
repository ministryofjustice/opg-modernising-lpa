package certificateproviderpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type completingYourIdentityConfirmationData struct {
	App      appcontext.Data
	Errors   validation.List
	Form     *howWillYouConfirmYourIdentityForm
	Options  howYouWillConfirmYourIdentityOptions
	Deadline time.Time
}

func CompletingYourIdentityConfirmation(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *certificateproviderdata.Provided) error {
		data := &completingYourIdentityConfirmationData{
			App:      appData,
			Form:     &howWillYouConfirmYourIdentityForm{},
			Options:  howYouWillConfirmYourIdentityValues,
			Deadline: provided.IdentityDeadline(),
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWillYouConfirmYourIdentityForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				return certificateprovider.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
