package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
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
	State page.TaskState
	Count int
}

func TaskList(tmpl template.Template, lpaStore LpaStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		tasks := getTasks(appData, lpa)

		var signPath string
		if tasks.ReadTheLpa.Completed() {
			ok, err := canSign(r.Context(), certificateProviderStore, lpa)
			if err != nil {
				return err
			}
			if ok {
				signPath = page.Paths.Attorney.Sign
			}
		}

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.Attorney.CheckYourName,
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:  "readTheLpa",
					Path:  page.Paths.Attorney.ReadTheLpa,
					State: tasks.ReadTheLpa,
				},
				{
					Name:  "signTheLpa",
					Path:  signPath,
					State: tasks.SignTheLpa,
				},
			},
		}

		return tmpl(w, data)
	}
}
