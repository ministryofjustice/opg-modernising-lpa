package voucherpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

func EnterAccessCode(voucherStore VoucherStore) page.EnterAccessCodeHandler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, session *sesh.LoginSession, lpa *lpadata.Lpa, shareCode sharecodedata.Link) error {
		if _, err := voucherStore.Create(r.Context(), shareCode, session.Email); err != nil {
			return fmt.Errorf("error creating voucher: %w", err)
		}

		return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
	}
}
