package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourChosenIdentityOptionsData struct {
	App            page.AppData
	Errors         validation.List
	IdentityOption identity.Option
	You            actor.Person
}

func YourChosenIdentityOptions(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, lpa, identityOptionPath(appData.Paths, lpa.IdentityOption))
		}

		data := &yourChosenIdentityOptionsData{
			App:            appData,
			IdentityOption: lpa.IdentityOption,
			You:            lpa.You,
		}

		return tmpl(w, data)
	}
}

func identityOptionPath(paths page.AppPaths, identityOption identity.Option) string {
	switch identityOption {
	case identity.OneLogin:
		return paths.IdentityWithOneLogin
	case identity.EasyID:
		return paths.IdentityWithYoti
	case identity.Passport:
		return paths.IdentityWithPassport
	case identity.BiometricResidencePermit:
		return paths.IdentityWithBiometricResidencePermit
	case identity.DrivingLicencePaper:
		return paths.IdentityWithDrivingLicencePaper
	case identity.DrivingLicencePhotocard:
		return paths.IdentityWithDrivingLicencePhotocard
	case identity.OnlineBankAccount:
		return paths.IdentityWithOnlineBankAccount
	default:
		panic("missing case in identityOptionPath")
	}
}
