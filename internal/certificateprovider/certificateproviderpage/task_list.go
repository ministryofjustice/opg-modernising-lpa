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
	App      appcontext.Data
	Errors   validation.List
	Provided *certificateproviderdata.Provided
	Lpa      *lpadata.Lpa
	Items    []taskListItem
}

type taskListItem struct {
	Name          string
	Path          certificateprovider.Path
	State         task.State
	IdentityState task.IdentityState
}

func TaskList(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		identityTaskPage := certificateprovider.PathConfirmYourIdentity
		switch provided.Tasks.ConfirmYourIdentity {
		case task.IdentityStateInProgress:
			identityTaskPage = certificateprovider.PathHowWillYouConfirmYourIdentity
		case task.IdentityStatePending:
			identityTaskPage = certificateprovider.PathCompletingYourIdentityConfirmation
		case task.IdentityStateCompleted:
			identityTaskPage = certificateprovider.PathReadTheLpa
		}

		confirmYourDetailsPage := certificateprovider.PathEnterDateOfBirth
		if provided.Tasks.ConfirmYourDetails.IsCompleted() {
			confirmYourDetailsPage = certificateprovider.PathConfirmYourDetails
		}

		data := &taskListData{
			App:      appData,
			Provided: provided,
			Lpa:      lpa,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  confirmYourDetailsPage,
					State: provided.Tasks.ConfirmYourDetails,
				},
				{
					Name:          "confirmYourIdentity",
					Path:          identityTaskPage,
					IdentityState: provided.Tasks.ConfirmYourIdentity,
				},
				{
					Name:  "provideYourCertificate",
					Path:  certificateprovider.PathReadTheLpa,
					State: provided.Tasks.ProvideTheCertificate,
				},
			},
		}

		return tmpl(w, data)
	}
}
