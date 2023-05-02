package attorney

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
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

		ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: lpa.ID})
		certificateProvider, err := certificateProviderStore.Get(ctx)
		if err != nil {
			if errors.Is(err, dynamo.NotFoundError{}) {
				certificateProvider = &actor.CertificateProvider{}
			} else {
				return err
			}
		}

		var signPath string
		progress := lpa.Progress(certificateProvider)

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
