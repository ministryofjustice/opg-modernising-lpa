package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *actor.Lpa
	Items  []taskListItem
}

type taskListItem struct {
	Name     string
	Path     string
	State    actor.TaskState
	Disabled bool
}

func TaskList(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		identityTaskPage := page.Paths.CertificateProvider.ProveYourIdentity
		if certificateProvider.CertificateProviderIdentityConfirmed(lpa.CertificateProvider.FirstNames, lpa.CertificateProvider.LastName) {
			identityTaskPage = page.Paths.CertificateProvider.ReadTheLpa
		}

		tasks := certificateProvider.Tasks

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.CertificateProvider.EnterDateOfBirth.Format(lpa.ID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:     "confirmYourIdentity",
					Path:     identityTaskPage.Format(lpa.ID),
					State:    tasks.ConfirmYourIdentity,
					Disabled: !lpa.Tasks.PayForLpa.IsCompleted() || lpa.SignedAt.IsZero(),
				},
				{
					Name:     "readTheLpa",
					Path:     page.Paths.CertificateProvider.ReadTheLpa.Format(lpa.ID),
					State:    tasks.ReadTheLpa,
					Disabled: lpa.SignedAt.IsZero(),
				},
				{
					Name:     "provideTheCertificateForThisLpa",
					Path:     page.Paths.CertificateProvider.ProvideCertificate.Format(lpa.ID),
					State:    tasks.ProvideTheCertificate,
					Disabled: lpa.SignedAt.IsZero() || !tasks.ConfirmYourDetails.Completed() || !tasks.ConfirmYourIdentity.Completed() || !tasks.ReadTheLpa.Completed(),
				},
			},
		}

		return tmpl(w, data)
	}
}
