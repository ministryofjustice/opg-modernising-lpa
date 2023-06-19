package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type yourChosenIdentityOptionsData struct {
	App            page.AppData
	Errors         validation.List
	IdentityOption identity.Option
	You            actor.Donor
}

func YourChosenIdentityOptions(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, lpa, identityOptionPath(appData.Paths, lpa.DonorIdentityOption, lpa.ID))
		}

		data := &yourChosenIdentityOptionsData{
			App:            appData,
			IdentityOption: lpa.DonorIdentityOption,
			You:            lpa.Donor,
		}

		return tmpl(w, data)
	}
}

func identityOptionPath(paths page.AppPaths, identityOption identity.Option, lpaID string) string {
	switch identityOption {
	case identity.OneLogin:
		return paths.IdentityWithOneLogin.Format()
	case identity.EasyID:
		return paths.IdentityWithYoti.Format(lpaID)
	case identity.Passport:
		return paths.IdentityWithPassport.Format(lpaID)
	case identity.BiometricResidencePermit:
		return paths.IdentityWithBiometricResidencePermit.Format(lpaID)
	case identity.DrivingLicencePaper:
		return paths.IdentityWithDrivingLicencePaper.Format(lpaID)
	case identity.DrivingLicencePhotocard:
		return paths.IdentityWithDrivingLicencePhotocard.Format(lpaID)
	case identity.OnlineBankAccount:
		return paths.IdentityWithOnlineBankAccount.Format(lpaID)
	default:
		panic("missing case in identityOptionPath")
	}
}
