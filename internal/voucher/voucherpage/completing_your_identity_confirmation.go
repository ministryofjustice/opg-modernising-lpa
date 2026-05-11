package voucherpage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type completingYourIdentityConfirmationData struct {
	App      appcontext.Data
	Errors   validation.List
	Form     *newforms.EnumForm[howYouWillConfirmYourIdentity, howYouWillConfirmYourIdentityOptions, *howYouWillConfirmYourIdentity]
	Donor    lpadata.Donor
	Deadline time.Time
}

func CompletingYourIdentityConfirmation(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		data := &completingYourIdentityConfirmationData{
			App:  appData,
			Form: newforms.NewEnumForm[howYouWillConfirmYourIdentity](appData.Localizer.T("howYouWouldLikeToContinue"), howYouWillConfirmYourIdentityValues),
		}

		if r.Method == http.MethodPost {
			if ok := data.Form.Parse(r); ok {
				return voucher.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
			}
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return fmt.Errorf("error retrieving lpa: %w", err)
		}

		data.Donor = lpa.Donor
		data.Deadline = provided.IdentityDeadline(lpa.SignedAt)

		return tmpl(w, data)
	}
}
