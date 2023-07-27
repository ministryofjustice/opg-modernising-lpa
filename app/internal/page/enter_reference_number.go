package page

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type enterReferenceNumberData struct {
	App    AppData
	Errors validation.List
	Form   *enterReferenceNumberForm
	Lpa    *Lpa
}

func EnterReferenceNumber(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore sessions.Store, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore, actorType actor.Type) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := data.Form.ReferenceNumber

				shareCode, err := shareCodeStore.Get(r.Context(), actorType, referenceNumber)
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

				ctx := ContextWithSessionData(r.Context(), &SessionData{
					SessionID: base64.StdEncoding.EncodeToString([]byte(session.Sub)),
					LpaID:     shareCode.LpaID,
				})

				redirect := Paths.CertificateProvider.WhoIsEligible.Format(shareCode.LpaID)

				if actorType.String() == actor.TypeCertificateProvider.String() {
					if _, err := certificateProviderStore.Create(ctx, shareCode.SessionID); err != nil {
						var ccf *types.ConditionalCheckFailedException
						if !errors.As(err, &ccf) {
							return err
						}
					}
				} else {
					if _, err := attorneyStore.Create(ctx, shareCode.SessionID, shareCode.AttorneyID, shareCode.IsReplacementAttorney); err != nil {
						var ccf *types.ConditionalCheckFailedException
						if !errors.As(err, &ccf) {
							return err
						}
					}

					redirect = Paths.Attorney.CodeOfConduct.Format(shareCode.LpaID)
				}

				appData.LpaID = shareCode.LpaID
				return appData.Redirect(w, r, nil, redirect)
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
		ReferenceNumber:    PostFormReferenceNumber(r, "reference-number"),
		ReferenceNumberRaw: PostFormString(r, "reference-number"),
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
