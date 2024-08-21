package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
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
	Name  string
	Path  string
	State task.State
	Count int
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *attorneydata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		tasks := provided.Tasks

		var signPath string
		if tasks.ConfirmYourDetails.IsCompleted() && tasks.ReadTheLpa.IsCompleted() &&
			!lpa.SignedAt.IsZero() && !lpa.CertificateProvider.SignedAt.IsZero() {
			signPath = attorney.PathRightsAndResponsibilities.Format(lpa.LpaID)
		}

		signItems := []taskListItem{{
			Name:  "signTheLpa",
			Path:  signPath,
			State: tasks.SignTheLpa,
		}}

		if provided.WouldLikeSecondSignatory.IsYes() && signPath != "" {
			signItems = []taskListItem{{
				Name:  "signTheLpaSignatory1",
				Path:  signPath,
				State: tasks.SignTheLpa,
			}, {
				Name:  "signTheLpaSignatory2",
				Path:  attorney.PathSign.Format(lpa.LpaID) + "?second",
				State: tasks.SignTheLpaSecond,
			}}
		}

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: append([]taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  attorney.PathPhoneNumber.Format(lpa.LpaID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:  "readTheLpa",
					Path:  attorney.PathReadTheLpa.Format(lpa.LpaID),
					State: tasks.ReadTheLpa,
				},
			}, signItems...),
		}

		return tmpl(w, data)
	}
}
