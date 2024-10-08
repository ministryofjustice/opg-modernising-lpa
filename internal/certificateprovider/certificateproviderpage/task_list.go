package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
	Items  []taskListItem
}

type taskListItem struct {
	Name     string
	Path     string
	State    task.State
	Disabled bool
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		identityTaskPage := certificateprovider.PathProveYourIdentity
		if certificateProvider.Tasks.ConfirmYourIdentity.IsCompleted() {
			identityTaskPage = certificateprovider.PathReadTheLpa
		}

		tasks := certificateProvider.Tasks

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  certificateprovider.PathEnterDateOfBirth.Format(lpa.LpaID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:     "confirmYourIdentity",
					Path:     identityTaskPage.Format(lpa.LpaID),
					State:    tasks.ConfirmYourIdentity,
					Disabled: !lpa.Paid || !lpa.SignedForDonor(),
				},
				{
					Name:     "provideYourCertificate",
					Path:     certificateprovider.PathReadTheLpa.Format(lpa.LpaID),
					State:    tasks.ProvideTheCertificate,
					Disabled: !lpa.SignedForDonor() || !tasks.ConfirmYourDetails.IsCompleted() || !tasks.ConfirmYourIdentity.IsCompleted(),
				},
			},
		}

		return tmpl(w, data)
	}
}
