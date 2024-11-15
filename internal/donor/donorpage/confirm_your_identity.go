package donorpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourIdentityData struct {
	App    appcontext.Data
	Errors validation.List
}

func ConfirmYourIdentity(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if provided.Tasks.ConfirmYourIdentity.IsNotStarted() {
				provided.Tasks.ConfirmYourIdentity = task.IdentityStateInProgress

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return fmt.Errorf("error updating donor: %w", err)
				}
			}

			return donor.PathIdentityWithOneLogin.Redirect(w, r, appData, provided)
		}

		return tmpl(w, &confirmYourIdentityData{App: appData})
	}
}
