package donorpage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App                            appcontext.Data
	Errors                         validation.List
	Donor                          *donordata.Provided
	DonorDetailsConfirmedByVoucher bool
}

func YourDetails(tmpl template.Template, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := yourDetailsData{
			App:   appData,
			Donor: provided,
		}

		if !provided.VoucherInvitedAt.IsZero() {
			voucher, err := voucherStore.GetAny(r.Context())
			if err != nil && !errors.As(err, &dynamo.NotFoundError{}) {
				return fmt.Errorf("error getting voucher: %w", err)
			}

			data.DonorDetailsConfirmedByVoucher = voucher != nil && voucher.DonorDetailsMatch.IsYes()
		}

		return tmpl(w, data)
	}
}
