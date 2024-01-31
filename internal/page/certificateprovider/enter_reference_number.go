package certificateprovider

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterReferenceNumberForm
	Donor  *actor.DonorProvidedDetails
}

func EnterReferenceNumber(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore sessions.Store, certificateProviderStore CertificateProviderStore) page.Handler {
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

				session, err := sesh.Login(sessionStore, r)
				if err != nil {
					return err
				}

				ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{
					SessionID: base64.StdEncoding.EncodeToString([]byte(session.Sub)),
					LpaID:     shareCode.LpaID,
				})

				if _, err := certificateProviderStore.Create(ctx, shareCode.SessionID); err != nil {
					var ccf *types.ConditionalCheckFailedException
					if !errors.As(err, &ccf) {
						return err
					}
				}

				if err := shareCodeStore.Delete(r.Context(), shareCode); err != nil {
					return err
				}

				appData.LpaID = shareCode.LpaID
				return page.Paths.CertificateProvider.WhoIsEligible.Redirect(w, r, appData, shareCode.LpaID)
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

	errors.String("reference-number", "theReferenceNumberYouEnter", f.ReferenceNumber,
		validation.StringLength(12))

	return errors
}
