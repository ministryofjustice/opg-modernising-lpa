package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type lpaProgressData struct {
	App        appcontext.Data
	Donor      *donordata.Provided
	Progress   task.Progress
	Lpa        *lpadata.Lpa
	Completed  []task.Step
	InProgress task.Step
	NotStarted []task.Step
	Errors     validation.List
}

func LpaProgress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &lpaProgressData{
			App:        appData,
			Donor:      donor,
			Completed:  donor.ProgressSteps.Completed(),
			InProgress: donor.ProgressSteps.InProgress(donor.FeeType.IsFullFee()),
			NotStarted: donor.ProgressSteps.RemainingDonorSteps(donor.FeeType.IsFullFee()),
			Lpa:        lpa,
		}

		return tmpl(w, data)
	}
}
