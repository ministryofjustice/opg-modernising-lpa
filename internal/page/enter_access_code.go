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

type EnterAccessCodeHandler func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, link accesscodedata.Link) error

type enterAccessCodeData struct {
	App  appcontext.Data
	Form *newforms.AccessCodeForm
}

func EnterAccessCode(tmpl template.Template, accessCodeStore AccessCodeStore, sessionStore UpdateLoginSessionStore, actorType actor.Type, next EnterAccessCodeHandler) Handler {
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

			if err := sessionStore.SetLogin(r, w, session); err != nil {
				return fmt.Errorf("saving login session: %w", err)
			}

			return next(appData, w, r.WithContext(ctx), session, accessCode)
		}

		return tmpl(w, data)
	}
}
