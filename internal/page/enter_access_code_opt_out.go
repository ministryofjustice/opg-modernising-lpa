package page

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type SetLpaDataSessionStore interface {
	SetLpaData(r *http.Request, w http.ResponseWriter, lpaDataSession *sesh.LpaDataSession) error
}

func EnterAccessCodeOptOut(tmpl template.Template, accessCodeStore AccessCodeStore, sessionStore SetLpaDataSessionStore, actorType actor.Type, redirect Path) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterAccessCodeData{
			App:  appData,
			Form: newforms.NewAccessCodeForm(appData.Localizer),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			referenceNumber := accesscodedata.HashedFromString(data.Form.AccessCode.Value, data.Form.DonorLastName.Value)

			accessCode, err := accessCodeStore.Get(r.Context(), actorType, referenceNumber)
			if err != nil {
				if errors.Is(err, dynamo.NotFoundError{}) {
					data.Form.AccessCode.Error = newforms.NewIncorrectError(appData.Localizer.T("accessCode"))
					data.Form.DonorLastName.Error = newforms.NewIncorrectError(appData.Localizer.T("donorLastName"))
					data.Form.Errors = append(data.Form.Errors, data.Form.AccessCode.Field, data.Form.DonorLastName.Field)
					return tmpl(w, data)
				}

				if errors.Is(err, dynamo.ErrTooManyRequests) {
					data.Form.AccessCode.Error = newforms.CustomError(appData.Localizer.T("tooManyAccessCodeAttempts"))
					data.Form.Errors = append(data.Form.Errors, data.Form.AccessCode.Field)
					return tmpl(w, data)
				}

				return fmt.Errorf("getting accesscode: %w", err)
			}

			if err := sessionStore.SetLpaData(r, w, &sesh.LpaDataSession{LpaID: accessCode.LpaKey.ID()}); err != nil {
				return fmt.Errorf("setting session lpa data: %w", err)
			}

			return redirect.RedirectQuery(w, r, appData, referenceNumber.Query())
		}

		return tmpl(w, data)
	}
}
