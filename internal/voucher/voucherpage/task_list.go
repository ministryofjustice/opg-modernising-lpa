package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type taskListData struct {
	App     appcontext.Data
	Errors  validation.List
	Voucher *voucherdata.Provided
	Items   []taskListItem
}

type taskListItem struct {
	Name  string
	Path  string
	State task.State
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		items := []taskListItem{
			{
				Name:  "confirmYourName",
				Path:  voucher.PathConfirmYourName.Format(appData.LpaID),
				State: provided.Tasks.ConfirmYourName,
			},
			{
				Name: appData.Localizer.Format("verifyPersonDetails", map[string]any{
					"DonorFullNamePossessive": appData.Localizer.Possessive(lpa.Donor.FullName()),
				}),
				Path:  voucher.PathVerifyDonorDetails.Format(appData.LpaID),
				State: provided.Tasks.VerifyDonorDetails,
			},
			{
				Name:  "confirmYourIdentity",
				Path:  voucher.PathConfirmYourIdentity.Format(appData.LpaID),
				State: provided.Tasks.ConfirmYourIdentity,
			},
			{
				Name:  "signTheDeclaration",
				Path:  voucher.PathSignTheDeclaration.Format(appData.LpaID),
				State: provided.Tasks.SignTheDeclaration,
			},
		}

		return tmpl(w, &taskListData{
			App:     appData,
			Voucher: provided,
			Items:   items,
		})
	}
}
