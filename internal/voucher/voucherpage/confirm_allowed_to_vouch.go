package voucherpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type confirmAllowedToVouchData struct {
	App                 appcontext.Data
	Errors              validation.List
	Form                *form.YesNoForm
	Lpa                 *lpadata.Lpa
	SurnameMatchesDonor bool
	MatchIdentity       bool
}

func ConfirmAllowedToVouch(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore, notifyClient NotifyClient, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &confirmAllowedToVouchData{
			App:                 appData,
			Form:                form.NewYesNoForm(form.YesNoUnknown),
			Lpa:                 lpa,
			SurnameMatchesDonor: strings.EqualFold(provided.LastName, lpa.Donor.LastName),
			MatchIdentity:       provided.Tasks.ConfirmYourIdentity.IsInProgress(),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfAllowedToVouch")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo.IsNo() {
					if err := notifyClient.SendActorEmail(r.Context(), lpa.Donor.Email, lpa.LpaUID, notify.VoucherFirstFailedVouchAttempt{
						Greeting:        notifyClient.EmailGreeting(lpa),
						VoucherFullName: provided.FullName(),
					}); err != nil {
						return err
					}

					donor, err := donorStore.GetAny(r.Context())
					if err != nil {
						return err
					}

					donor.FailedVouchAttempts++

					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}

					return voucher.PathYouCannotVouchForDonor.Redirect(w, r, appData, appData.LpaID)
				}

				if provided.Tasks.ConfirmYourIdentity.IsInProgress() {
					provided.Tasks.ConfirmYourIdentity = task.StateCompleted
				} else {
					provided.Tasks.ConfirmYourName = task.StateCompleted
				}

				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
