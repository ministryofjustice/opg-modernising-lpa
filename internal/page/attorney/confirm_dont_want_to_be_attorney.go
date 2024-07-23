package attorney

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeAttorneyData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *lpastore.Lpa
}

func ConfirmDontWantToBeAttorney(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, attorneyStore AttorneyStore, notifyClient NotifyClient, appPublicURL string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeAttorneyData{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			attorneyFullName, err := findAttorneyFullName(lpa, attorneyProvidedDetails.UID, attorneyProvidedDetails.IsTrustCorporation, attorneyProvidedDetails.IsReplacement)
			if err != nil {
				return err
			}

			email := notify.AttorneyOptedOutEmail{
				AttorneyFullName:  attorneyFullName,
				DonorFullName:     lpa.Donor.FullName(),
				LpaType:           appData.Localizer.T(lpa.Type.String()),
				LpaUID:            lpa.LpaUID,
				DonorStartPageURL: appPublicURL + page.Paths.Start.Format(),
			}

			if err := attorneyStore.Delete(r.Context()); err != nil {
				return err
			}

			if err := notifyClient.SendActorEmail(r.Context(), lpa.Donor.Email, lpa.LpaUID, email); err != nil {
				return err
			}

			return page.Paths.Attorney.YouHaveDecidedNotToBeAttorney.RedirectQuery(w, r, appData, url.Values{
				"donorFullName":   {lpa.Donor.FullName()},
				"donorFirstNames": {lpa.Donor.FirstNames},
			})
		}

		return tmpl(w, data)
	}
}
