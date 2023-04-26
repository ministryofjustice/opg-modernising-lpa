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

func TaskList(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		var signPath string
		progress := lpa.Progress()

		if progress.LpaSigned.Completed() && progress.CertificateProviderDeclared.Completed() {
			signPath = page.Paths.Attorney.Sign
		}

		tasks := getTasks(appData, lpa)

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
					Path:  page.Paths.Attorney.NextPage,
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
