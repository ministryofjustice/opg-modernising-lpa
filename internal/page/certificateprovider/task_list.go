package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *lpastore.Lpa
	Items  []taskListItem
}

type taskListItem struct {
	Name     string
	Path     string
	State    actor.TaskState
	Disabled bool
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		identityTaskPage := page.Paths.CertificateProvider.ProveYourIdentity
		if certificateProvider.Tasks.ConfirmYourIdentity.Completed() {
			identityTaskPage = page.Paths.CertificateProvider.ReadTheLpa
		}

		tasks := certificateProvider.Tasks

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.CertificateProvider.EnterDateOfBirth.Format(lpa.LpaID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:     "confirmYourIdentity",
					Path:     identityTaskPage.Format(lpa.LpaID),
					State:    tasks.ConfirmYourIdentity,
					Disabled: !lpa.Paid || lpa.SignedAt.IsZero(),
				},
				{
					Name:     "provideYourCertificate",
					Path:     page.Paths.CertificateProvider.ReadTheLpa.Format(lpa.LpaID),
					State:    tasks.ProvideTheCertificate,
					Disabled: lpa.SignedAt.IsZero() || !tasks.ConfirmYourDetails.Completed() || !tasks.ConfirmYourIdentity.Completed(),
				},
			},
		}

		return tmpl(w, data)
	}
}
