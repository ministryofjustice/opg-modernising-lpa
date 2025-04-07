package certificateproviderpage

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumberData struct {
	App          appcontext.Data
	Errors       validation.List
	Form         *enterReferenceNumberForm
	HideLoginNav bool
}

func EnterReferenceNumber(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore SessionStore, certificateProviderStore CertificateProviderStore, lpaStoreClient LpaStoreClient, dashboardStore DashboardStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		results, err := dashboardStore.GetAll(r.Context())
		if err != nil {
			return fmt.Errorf("error getting dashboard results: %w", err)
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := sharecodedata.HashedFromString(data.Form.ReferenceNumber)

				shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeCertificateProvider, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"})
						return tmpl(w, data)
					} else {
						return fmt.Errorf("error getting shareCode: %w", err)
					}
				}

				lpa, err := lpaStoreClient.Lpa(r.Context(), shareCode.LpaUID)
				if err != nil && !errors.Is(err, lpastore.ErrNotFound) {
					return fmt.Errorf("error getting LPA from LPA store: %w", err)
				}

				if lpa != nil && lpa.CertificateProvider.Channel.IsPaper() && !lpa.CertificateProvider.SignedAt.IsZero() {
					redirectTo := page.PathCertificateProviderYouHaveAlreadyProvidedACertificateLoggedIn

					if results.Empty() {
						if err = sessionStore.ClearLogin(r, w); err != nil {
							return fmt.Errorf("error clearing login session: %w", err)
						}

						redirectTo = page.PathCertificateProviderYouHaveAlreadyProvidedACertificate
					}

					return redirectTo.RedirectQuery(w, r, appData, url.Values{
						"donorFullName": {lpa.Donor.FullName()},
						"lpaType":       {appData.Localizer.T(lpa.Type.String())},
					})
				}

				session, err := sessionStore.Login(r)
				if err != nil {
					return fmt.Errorf("error setting login session: %w", err)
				}

				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{
					SessionID: session.SessionID(),
					LpaID:     shareCode.LpaKey.ID(),
				})

				if _, err := certificateProviderStore.Create(ctx, shareCode, session.Email); err != nil {
					return fmt.Errorf("error creating certificate provider: %w", err)
				}

				appData.LpaID = shareCode.LpaKey.ID()
				return certificateprovider.PathWhoIsEligible.Redirect(w, r, appData, appData.LpaID)
			}
		}

		if results.Empty() {
			data.App.SessionID = ""
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
