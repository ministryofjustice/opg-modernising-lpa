package supporterpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type viewLPAData struct {
	App      appcontext.Data
	Errors   validation.List
	Lpa      *lpastore.Lpa
	Progress page.Progress
}

func ViewLPA(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
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
