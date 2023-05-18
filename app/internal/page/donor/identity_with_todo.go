package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityWithTodoData struct {
	App            page.AppData
	Errors         validation.List
	IdentityOption identity.Option
}

func IdentityWithTodo(tmpl template.Template, donorStore DonorStore, now func() time.Time, identityOption identity.Option) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, lpa, page.Paths.ReadYourLpa)
		}

		lpa.DonorIdentityUserData = identity.UserData{
			OK:          true,
			Provider:    identityOption,
			FirstNames:  lpa.Donor.FirstNames,
			LastName:    lpa.Donor.LastName,
			DateOfBirth: lpa.Donor.DateOfBirth,
			RetrievedAt: now(),
		}
		if err := donorStore.Put(r.Context(), lpa); err != nil {
			return err
		}

		data := &identityWithTodoData{
			App:            appData,
			IdentityOption: identityOption,
		}

		return tmpl(w, data)
	}
}
