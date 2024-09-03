package attorneypage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeAttorneyData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ConfirmDontWantToBeAttorney(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, attorneyStore AttorneyStore, notifyClient NotifyClient, appPublicURL string, lpaStoreClient LpaStoreClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeAttorneyData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			fullName, actorType := lpa.Attorney(attorneyProvidedDetails.UID)
			if actorType.IsNone() {
				return errors.New("attorney not found")
			}

			email := notify.AttorneyOptedOutEmail{
				Greeting:          notifyClient.EmailGreeting(lpa),
				AttorneyFullName:  fullName,
				DonorFullName:     lpa.Donor.FullName(),
				LpaType:           appData.Localizer.T(lpa.Type.String()),
				LpaUID:            lpa.LpaUID,
				DonorStartPageURL: appPublicURL + page.PathStart.Format(),
			}

			if err := notifyClient.SendActorEmail(r.Context(), lpa.CorrespondentEmail(), lpa.LpaUID, email); err != nil {
				return err
			}

			if err := lpaStoreClient.SendAttorneyOptOut(r.Context(), lpa.LpaUID, attorneyProvidedDetails.UID, actorType); err != nil {
				return err
			}

			if err := attorneyStore.Delete(r.Context()); err != nil {
				return err
			}

			return page.PathAttorneyYouHaveDecidedNotToBeAttorney.RedirectQuery(w, r, appData, url.Values{
				"donorFullName":   {lpa.Donor.FullName()},
				"donorFirstNames": {lpa.Donor.FirstNames},
			})
		}

		return tmpl(w, data)
	}
}
