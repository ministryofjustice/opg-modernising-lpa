package voucherpage

import (
	"cmp"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type confirmYourNameData struct {
	App        appcontext.Data
	Errors     validation.List
	Lpa        *lpadata.Lpa
	Tasks      voucherdata.Tasks
	FirstNames string
	LastName   string
	Changed    bool
	ShowBanner bool
}

func ConfirmYourName(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		firstNames := cmp.Or(provided.FirstNames, lpa.Voucher.FirstNames)
		lastName := cmp.Or(provided.LastName, lpa.Voucher.LastName)

		if r.Method == http.MethodPost {
			redirect := voucher.PathTaskList
			state := task.StateCompleted

			provided.FirstNames = firstNames
			provided.LastName = lastName

			if !provided.Tasks.ConfirmYourName.IsCompleted() && !provided.NameMatches(lpa).IsNone() {
				redirect = voucher.PathConfirmAllowedToVouch
				state = task.StateInProgress
			}

			provided.Tasks.ConfirmYourName = state
			if err := voucherStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return redirect.Redirect(w, r, appData, appData.LpaID)
		}

		cookie, _ := r.Cookie("banner")
		if cookie != nil {
			cookie.MaxAge = -1
			http.SetCookie(w, cookie)
		}

		return tmpl(w, &confirmYourNameData{
			App:        appData,
			Lpa:        lpa,
			Tasks:      provided.Tasks,
			FirstNames: firstNames,
			LastName:   lastName,
			Changed: (provided.FirstNames != lpa.Voucher.FirstNames || provided.LastName != lpa.Voucher.LastName) &&
				(provided.FirstNames != "" || provided.LastName != ""),
			ShowBanner: cookie != nil && cookie.Value == "1",
		})
	}
}
