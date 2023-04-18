package attorney

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterReferenceNumberForm
	Lpa    *page.Lpa
}

func EnterReferenceNumber(tmpl template.Template, lpaStore LpaStore, dataStore DataStore, sessionStore SessionStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := data.Form.ReferenceNumber

				var v page.ShareCodeData
				if err := dataStore.Get(r.Context(), "ATTORNEYSHARE#"+referenceNumber, "#METADATA#"+referenceNumber, &v); err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
						return tmpl(w, data)
					} else {
						return err
					}
				}

				session, err := sesh.Attorney(sessionStore, r)
				if err != nil {
					return err
				}
				session.LpaID = v.LpaID
				session.DonorSessionID = v.SessionID

				if err := sesh.SetAttorney(sessionStore, r, w, session); err != nil {
					return err
				}

				lpa, err := lpaStore.Get(page.ContextWithSessionData(r.Context(), &page.SessionData{
					SessionID: v.SessionID,
					LpaID:     v.LpaID,
				}))
				if err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.Attorney.DateOfBirth)
			}
		}

		return tmpl(w, data)
	}
}

type enterReferenceNumberForm struct {
	ReferenceNumber string
}

func (f *enterReferenceNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-number", "twelveCharactersAttorneyReferenceNumber", strings.ReplaceAll(f.ReferenceNumber, " ", ""),
		validation.Empty(),
	)

	errors.String("reference-number", "attorneyReferenceNumberMustBeTwelveCharacters", strings.ReplaceAll(f.ReferenceNumber, " ", ""),
		validation.StringLength(12),
	)

	return errors
}

func readEnterReferenceNumberForm(r *http.Request) *enterReferenceNumberForm {
	return &enterReferenceNumberForm{
		ReferenceNumber: page.PostFormString(r, "reference-number"),
	}
}
