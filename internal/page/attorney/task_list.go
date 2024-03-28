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
	Donor  *actor.DonorProvidedDetails
	Items  []taskListItem
}

type taskListItem struct {
	Name  string
	Path  string
	State actor.TaskState
	Count int
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorney *actor.AttorneyProvidedDetails) error {
		donor, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		tasks := attorney.Tasks

		var signPath string
		if tasks.ConfirmYourDetails.Completed() && tasks.ReadTheLpa.Completed() {
			ok, err := canSign(r.Context(), certificateProviderStore, donor)
			if err != nil {
				return err
			}
			if ok {
				signPath = page.Paths.Attorney.Sign.Format(donor.LpaID)
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
			App:   appData,
			Donor: donor,
			Items: append([]taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.Attorney.MobileNumber.Format(donor.LpaID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:  "readTheLpa",
					Path:  page.Paths.Attorney.ReadTheLpa.Format(donor.LpaID),
					State: tasks.ReadTheLpa,
				},
			}, signItems...),
		}

		return tmpl(w, data)
	}
}
