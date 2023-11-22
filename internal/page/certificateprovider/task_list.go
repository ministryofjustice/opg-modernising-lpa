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
	Donor  *actor.DonorProvidedDetails
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
		donor, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		identityTaskPage := page.Paths.CertificateProvider.ProveYourIdentity
		if certificateProvider.CertificateProviderIdentityConfirmed(donor.CertificateProvider.FirstNames, donor.CertificateProvider.LastName) {
			identityTaskPage = page.Paths.CertificateProvider.ReadTheLpa
		}

		tasks := certificateProvider.Tasks

		data := &taskListData{
			App:   appData,
			Donor: donor,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.CertificateProvider.EnterDateOfBirth.Format(donor.LpaID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:     "confirmYourIdentity",
					Path:     identityTaskPage.Format(donor.LpaID),
					State:    tasks.ConfirmYourIdentity,
					Disabled: !donor.Tasks.PayForLpa.IsCompleted() || donor.SignedAt.IsZero(),
				},
				{
					Name:     "provideYourCertificate",
					Path:     page.Paths.CertificateProvider.ReadTheLpa.Format(donor.LpaID),
					State:    tasks.ProvideTheCertificate,
					Disabled: donor.SignedAt.IsZero() || !tasks.ConfirmYourDetails.Completed() || !tasks.ConfirmYourIdentity.Completed(),
				},
			},
		}

		return tmpl(w, data)
	}
}
