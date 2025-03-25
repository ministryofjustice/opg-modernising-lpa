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
	Name          string
	Path          voucher.Path
	State         task.State
	IdentityState task.IdentityState
}

func TaskList(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		confirmYourIdentityPath := voucher.PathConfirmYourIdentity
		switch provided.Tasks.ConfirmYourIdentity {
		case task.IdentityStateInProgress:
			if provided.IdentityUserData.CheckedAt.IsZero() {
				confirmYourIdentityPath = voucher.PathHowWillYouConfirmYourIdentity
			} else {
				confirmYourIdentityPath = voucher.PathConfirmAllowedToVouch
			}
		case task.IdentityStatePending:
			confirmYourIdentityPath = voucher.PathCompletingYourIdentityConfirmation
		case task.IdentityStateCompleted:
			confirmYourIdentityPath = voucher.PathOneLoginIdentityDetails
		}

		items := []taskListItem{
			{
				Name:  "confirmYourName",
				Path:  voucher.PathConfirmYourName,
				State: provided.Tasks.ConfirmYourName,
			},
			{
				Name: appData.Localizer.Format("verifyPersonIdentity", map[string]any{
					"DonorFullName": lpa.Donor.FullName(),
				}),
				Path:  voucher.PathVerifyDonorDetails,
				State: provided.Tasks.VerifyDonorDetails,
			},
			{
				Name:          "confirmYourIdentity",
				Path:          confirmYourIdentityPath,
				IdentityState: provided.Tasks.ConfirmYourIdentity,
			},
			{
				Name:  "signTheDeclaration",
				Path:  voucher.PathSignTheDeclaration,
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
