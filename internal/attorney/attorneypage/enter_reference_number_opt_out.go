package attorneypage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

func EnterReferenceNumberOptOut(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore SessionStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				referenceNumber := sharecodedata.HashedFromString(data.Form.ReferenceNumber)

				shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeAttorney, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
						return tmpl(w, data)
					} else {
						return err
					}
				}

				if err := sessionStore.SetLpaData(r, w, &sesh.LpaDataSession{LpaID: shareCode.LpaKey.ID()}); err != nil {
					return err
				}

				return page.PathAttorneyConfirmDontWantToBeAttorneyLoggedOut.RedirectQuery(w, r, appData, referenceNumber.Query())
			}
		}

		return tmpl(w, data)
	}
}
