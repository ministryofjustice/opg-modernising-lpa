package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaProgressData struct {
	App      page.AppData
	Donor    *donordata.Provided
	Progress page.Progress
	Errors   validation.List
}

func LpaProgress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &lpaProgressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(lpa),
		}

		return tmpl(w, data)
	}
}
