package certificateproviderpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type completingYourIdentityConfirmationData struct {
	App      appcontext.Data
	Errors   validation.List
	Form     *form.SelectForm[howYouWillConfirmYourIdentity, howYouWillConfirmYourIdentityOptions, *howYouWillConfirmYourIdentity]
	Deadline time.Time
}

func CompletingYourIdentityConfirmation(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *certificateproviderdata.Provided, _ *lpadata.Lpa) error {
		data := &completingYourIdentityConfirmationData{
			App:      appData,
			Form:     form.NewEmptySelectForm[howYouWillConfirmYourIdentity](howYouWillConfirmYourIdentityValues, "howYouWouldLikeToContinue"),
			Deadline: provided.IdentityDeadline(),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				return certificateprovider.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
