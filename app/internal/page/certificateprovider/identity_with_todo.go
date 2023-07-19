package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type identityWithTodoData struct {
	App            page.AppData
	Errors         validation.List
	IdentityOption identity.Option
}

func IdentityWithTodo(tmpl template.Template, now func() time.Time, identityOption identity.Option, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.ReadTheLpa.Format(certificateProvider.LpaID))
		}

		certificateProvider.IdentityUserData = identity.UserData{
			OK:          true,
			Provider:    identityOption,
			FirstNames:  certificateProvider.FirstNames,
			LastName:    certificateProvider.LastName,
			DateOfBirth: certificateProvider.DateOfBirth,
			RetrievedAt: now(),
		}
		certificateProvider.Tasks.ConfirmYourIdentity = actor.TaskCompleted

		if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
			return err
		}

		data := &identityWithTodoData{
			App:            appData,
			IdentityOption: identityOption,
		}

		return tmpl(w, data)
	}
}
