package attorney

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
	Lpa    *page.Lpa
	Items  []taskListItem
}

type taskListItem struct {
	Name  string
	Path  string
	State actor.TaskState
	Count int
}

func TaskList(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorney *actor.AttorneyProvidedDetails) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		tasks := attorney.Tasks

		var signPath string
		if tasks.ConfirmYourDetails.Completed() && tasks.ReadTheLpa.Completed() {
			ok, err := canSign(r.Context(), certificateProviderStore, lpa)
			if err != nil {
				return err
			}
			if ok {
				signPath = page.Paths.Attorney.Sign.Format(lpa.ID)
			}
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
				Path:  signPath + "?second",
				State: tasks.SignTheLpaSecond,
			}}
		}

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: append([]taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.Attorney.MobileNumber.Format(lpa.ID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:  "readTheLpa",
					Path:  page.Paths.Attorney.ReadTheLpa.Format(lpa.ID),
					State: tasks.ReadTheLpa,
				},
			}, signItems...),
		}

		return tmpl(w, data)
	}
}
