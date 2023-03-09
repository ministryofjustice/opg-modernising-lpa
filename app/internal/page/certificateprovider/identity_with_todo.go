package certificateprovider

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

func IdentityWithTodo(tmpl template.Template, lpaStore LpaStore, now func() time.Time, identityOption identity.Option) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, nil, page.Paths.CertificateProviderReadTheLpa)
		}

		lpa.CertificateProviderIdentityUserData = identity.UserData{
			OK:          true,
			Provider:    identityOption,
			FirstNames:  lpa.CertificateProvider.FirstNames,
			LastName:    lpa.CertificateProvider.LastName,
			RetrievedAt: now(),
		}
		if err := lpaStore.Put(r.Context(), lpa); err != nil {
			return err
		}

		data := &identityWithTodoData{
			App:            appData,
			IdentityOption: identityOption,
		}

		return tmpl(w, data)
	}
}
