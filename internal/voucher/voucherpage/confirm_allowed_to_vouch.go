package voucherpage

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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

func ConfirmAllowedToVouch(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore, fail vouchFailer) Handler {
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
					if err := fail(r.Context(), provided, lpa); err != nil {
						return err
					}

					return page.PathYouCannotVouchForDonor.RedirectQuery(w, r, appData, url.Values{
						"donorFullName": {lpa.Donor.FullName()},
					})
				}

				if provided.Tasks.ConfirmYourIdentity.IsInProgress() {
					provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
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
