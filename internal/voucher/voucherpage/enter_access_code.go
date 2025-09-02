package voucherpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

func EnterAccessCode(voucherStore VoucherStore) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, link accesscodedata.Link) error {
		if _, err := voucherStore.Create(r.Context(), link, session.Email); err != nil {
			return fmt.Errorf("error creating voucher: %w", err)
		}

		return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
	}
}
