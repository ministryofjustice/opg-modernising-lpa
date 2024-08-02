package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpastore.Lpa
	Items  []taskListItem
}

type taskListItem struct {
	Name  string
	Path  string
	State task.State
	Count int
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorney *attorneydata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		tasks := attorney.Tasks

		var signPath string
		if tasks.ConfirmYourDetails.Completed() && tasks.ReadTheLpa.Completed() &&
			!lpa.SignedAt.IsZero() && !lpa.CertificateProvider.SignedAt.IsZero() {
			signPath = page.Paths.Attorney.RightsAndResponsibilities.Format(lpa.LpaID)
		}

		signItems := []taskListItem{{
			Name:  "signTheLpa",
			Path:  signPath,
			State: tasks.SignTheLpa,
		}}

		if attorney.WouldLikeSecondSignatory.IsYes() && signPath != "" {
			signItems = []taskListItem{{
				Name:  "signTheLpaSignatory1",
				Path:  signPath,
				State: tasks.SignTheLpa,
			}, {
				Name:  "signTheLpaSignatory2",
				Path:  page.Paths.Attorney.Sign.Format(lpa.LpaID) + "?second",
				State: tasks.SignTheLpaSecond,
			}}
		}

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: append([]taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.Attorney.MobileNumber.Format(lpa.LpaID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:  "readTheLpa",
					Path:  page.Paths.Attorney.ReadTheLpa.Format(lpa.LpaID),
					State: tasks.ReadTheLpa,
				},
			}, signItems...),
		}

		return tmpl(w, data)
	}
}
