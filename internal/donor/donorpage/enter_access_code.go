package donorpage

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterAccessCodeData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.AccessCodeForm
}

func EnterAccessCode(logger Logger, tmpl template.Template, accessCodeStore AccessCodeStore, sessionStore SessionStore, donorStore DonorStore, eventClient EventClient) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := enterAccessCodeData{
			App:  appData,
			Form: form.NewAccessCodeForm(),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				referenceNumber := accesscodedata.HashedFromString(data.Form.AccessCode, data.Form.DonorLastName)

				accessCode, err := accessCodeStore.Get(r.Context(), actor.TypeDonor, referenceNumber)
				if err != nil {
					if errors.Is(err, dynamo.NotFoundError{}) {
						data.Errors.Add(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"})
						data.Errors.Add(form.FieldNames.DonorLastName, validation.IncorrectError{Label: "donorLastName"})
						return tmpl(w, data)
					}

					return fmt.Errorf("get accesscode: %w", err)
				}

				session, err := sessionStore.Login(r)
				if err != nil {
					return fmt.Errorf("getting login session: %w", err)
				}

				session.HasLPAs = true

				if err := sessionStore.SetLogin(r, w, session); err != nil {
					return fmt.Errorf("saving login session: %w", err)
				}

				appSession := &appcontext.Session{
					SessionID: session.SessionID(),
					LpaID:     accessCode.LpaKey.ID(),
				}
				if org, ok := accessCode.LpaOwnerKey.Organisation(); ok {
					appSession.OrganisationID = org.ID()
				}

				r = r.WithContext(appcontext.ContextWithSession(r.Context(), appSession))
				appData.LpaID = accessCode.LpaKey.ID()

				if err := donorStore.Link(r.Context(), accessCode, session.Email); err != nil {
					return fmt.Errorf("link donor: %w", err)
				}

				logger.InfoContext(r.Context(), "donor access added", slog.String("lpa_id", accessCode.LpaKey.ID()))

				if err := eventClient.SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineDonor); err != nil {
					return fmt.Errorf("sending metric: %w", err)
				}

				return page.PathDashboard.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}
