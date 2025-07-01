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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type SetLpaDataSessionStore interface {
	SetLpaData(r *http.Request, w http.ResponseWriter, lpaDataSession *sesh.LpaDataSession) error
}

func EnterAccessCodeOptOut(tmpl template.Template, accessCodeStore AccessCodeStore, sessionStore SetLpaDataSessionStore, lpaStoreResolvingService LpaStoreResolvingService, actorType actor.Type, redirect Path) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterAccessCodeData{
			App:  appData,
			Form: form.NewAccessCodeForm(),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				referenceNumber := accesscodedata.HashedFromString(data.Form.AccessCode)

				accessCode, err := accessCodeStore.Get(r.Context(), actorType, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"})
						return tmpl(w, data)
					} else {
						return fmt.Errorf("getting accesscode: %w", err)
					}
				}

				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{
					LpaID: accessCode.LpaKey.ID(),
				})

				lpa, err := lpaStoreResolvingService.Get(ctx)
				if err != nil {
					return fmt.Errorf("resolving lpa: %w", err)
				}

				if lpa.Donor.LastName != data.Form.DonorLastName {
					data.Errors.Add(form.FieldNames.DonorLastName, validation.IncorrectError{Label: "donorLastName"})
					return tmpl(w, data)
				}

				if err := sessionStore.SetLpaData(r, w, &sesh.LpaDataSession{LpaID: accessCode.LpaKey.ID()}); err != nil {
					return fmt.Errorf("setting session lpa data: %w", err)
				}

				return redirect.RedirectQuery(w, r, appData, referenceNumber.Query())
			}
		}

		return tmpl(w, data)
	}
}
