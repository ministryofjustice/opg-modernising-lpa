package supporterpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type viewLPAData struct {
	App      appcontext.Data
	Errors   validation.List
	Lpa      *lpadata.Lpa
	Progress task.Progress
}

func ViewLPA(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, &viewLPAData{
			App:      appData,
			Lpa:      lpa,
			Progress: progressTracker.Progress(lpa),
		})
	}
}
