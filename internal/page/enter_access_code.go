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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type EnterAccessCodeHandler func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, lpa *lpadata.Lpa, link accesscodedata.Link) error

type enterAccessCodeData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.AccessCodeForm
}

func EnterAccessCode(tmpl template.Template, accessCodeStore AccessCodeStore, sessionStore UpdateLoginSessionStore, lpaStoreResolvingService LpaStoreResolvingService, actorType actor.Type, next EnterAccessCodeHandler) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterAccessCodeData{
			App:  appData,
			Form: form.NewAccessCodeForm(),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := accesscodedata.HashedFromString(data.Form.AccessCode)

				accessCode, err := accessCodeStore.Get(r.Context(), actorType, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"})
						return tmpl(w, data)
					}

					return err
				}

				session, err := sessionStore.Login(r)
				if err != nil {
					return fmt.Errorf("getting login session: %w", err)
				}

				session.HasLPAs = true

				appSession := &appcontext.Session{
					SessionID: session.SessionID(),
					LpaID:     accessCode.LpaKey.ID(),
				}
				if org, ok := accessCode.LpaOwnerKey.Organisation(); ok {
					appSession.OrganisationID = org.ID()
				}

				ctx := appcontext.ContextWithSession(r.Context(), appSession)
				appData.LpaID = accessCode.LpaKey.ID()

				lpa, err := lpaStoreResolvingService.Get(ctx)
				if err != nil {
					return fmt.Errorf("getting LPA from LPA store: %w", err)
				}

				if lpa.Donor.LastName != data.Form.DonorLastName {
					data.Errors.Add(form.FieldNames.DonorLastName, validation.IncorrectError{Label: "donorLastName"})
					return tmpl(w, data)
				}

				if err := sessionStore.SetLogin(r, w, session); err != nil {
					return fmt.Errorf("saving login session: %w", err)
				}

				return next(appData, w, r.WithContext(ctx), session, lpa, accessCode)
			}
		}

		return tmpl(w, data)
	}
}
