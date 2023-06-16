package certificateprovider

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type enterReferenceNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterReferenceNumberForm
	Lpa    *page.Lpa
}

func EnterReferenceNumber(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore sessions.Store) page.Handler {
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

				shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeCertificateProvider, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
						return tmpl(w, data)
					} else {
						return err
					}
				}

				if err := sesh.SetShareCode(sessionStore, r, w, &sesh.ShareCodeSession{
					LpaID:           shareCode.LpaID,
					SessionID:       shareCode.SessionID,
					Identity:        shareCode.Identity,
					DonorFullName:   shareCode.DonorFullname,
					DonorFirstNames: shareCode.DonorFirstNames,
				}); err != nil {
					return err
				}

				appData.Redirect(w, r, nil, page.Paths.CertificateProvider.WhoIsEligible)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type enterReferenceNumberForm struct {
	ReferenceNumber    string
	ReferenceNumberRaw string
}

func readEnterReferenceNumberForm(r *http.Request) *enterReferenceNumberForm {
	return &enterReferenceNumberForm{
		ReferenceNumber:    page.PostFormReferenceNumber(r, "reference-number"),
		ReferenceNumberRaw: page.PostFormString(r, "reference-number"),
	}
}

func (f *enterReferenceNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("reference-number", "twelveCharactersReferenceNumber", f.ReferenceNumber,
		validation.Empty())

	errors.String("reference-number", "referenceNumberMustBeTwelveCharacters", f.ReferenceNumber,
		validation.StringLength(12))

	return errors
}
