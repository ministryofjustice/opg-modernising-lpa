package attorneypage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterReferenceNumberData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterReferenceNumberForm
}

func EnterReferenceNumber(tmpl template.Template, shareCodeStore ShareCodeStore, sessionStore SessionStore, attorneyStore AttorneyStore, lpaStoreClient LpaStoreClient) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterReferenceNumberData{
			App:  appData,
			Form: &enterReferenceNumberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterReferenceNumberForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
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

				lpa, err := lpaStoreClient.Lpa(r.Context(), shareCode.LpaUID)
				if err != nil && !errors.Is(err, lpastore.ErrNotFound) {
					return fmt.Errorf("error getting LPA from LPA store: %w", err)
				}

				if lpa != nil {
					lpaAttorney, found := lpa.Attorneys.Get(shareCode.ActorUID)

					if found && lpaAttorney.Channel.IsPaper() && !lpaAttorney.SignedAt.IsZero() {
						if err := lpaStoreClient.SendPaperAttorneyAccessOnline(r.Context(), shareCode.LpaUID, appData.LoginSessionEmail, shareCode.ActorUID); err != nil {
							return fmt.Errorf("error sending attorney email to LPA store: %w", err)
						}

						return page.PathDashboard.Redirect(w, r, appData)
					}
				}

				session, err := sessionStore.Login(r)
				if err != nil {
					return err
				}

				ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{
					SessionID: session.SessionID(),
					LpaID:     shareCode.LpaKey.ID(),
				})

				if _, err := attorneyStore.Create(ctx, shareCode, session.Email); err != nil {
					return err
				}

				appData.LpaID = shareCode.LpaKey.ID()
				return attorney.PathCodeOfConduct.Redirect(w, r, appData, appData.LpaID)
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
