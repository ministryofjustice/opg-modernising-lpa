package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourChosenIdentityOptionsData struct {
	App            AppData
	Errors         validation.List
	IdentityOption IdentityOption
	You            actor.Person
}

func YourChosenIdentityOptions(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
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

func identityOptionPath(paths AppPaths, identityOption IdentityOption) string {
	switch identityOption {
	case OneLogin:
		return paths.IdentityWithOneLogin
	case EasyID:
		return paths.IdentityWithYoti
	case Passport:
		return paths.IdentityWithPassport
	case BiometricResidencePermit:
		return paths.IdentityWithBiometricResidencePermit
	case DrivingLicencePaper:
		return paths.IdentityWithDrivingLicencePaper
	case DrivingLicencePhotocard:
		return paths.IdentityWithDrivingLicencePhotocard
	case OnlineBankAccount:
		return paths.IdentityWithOnlineBankAccount
	default:
		panic("missing case in identityOptionPath")
	}
}
