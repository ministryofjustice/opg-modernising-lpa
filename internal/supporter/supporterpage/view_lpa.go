package supporterpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/progress"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type viewLPAData struct {
	App         appcontext.Data
	Errors      validation.List
	Lpa         *lpadata.Lpa
	Completed   []progress.Step
	InProgress  progress.Step
	NotStarted  []progress.Step
	IsSupporter bool
}

func ViewLPA(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		donor, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		progressTracker.Init(donor.FeeType.IsFullFee(), lpa.IsOrganisationDonor, donor.ProgressSteps.CompletedSteps)
		inProgress, notStarted := progressTracker.Remaining()

		return tmpl(w, &viewLPAData{
			App:         appData,
			Lpa:         lpa,
			Completed:   progressTracker.Completed(),
			InProgress:  inProgress,
			NotStarted:  notStarted,
			IsSupporter: progressTracker.IsSupporter(),
		})
	}
}
