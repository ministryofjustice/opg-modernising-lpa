package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourChosenIdentityOptionsData struct {
	App            page.AppData
	Errors         validation.List
	IdentityOption identity.Option
}

func YourChosenIdentityOptions(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, lpa, identityOptionPath(appData.Paths, lpa.CertificateProviderIdentityOption))
		}

		data := &yourChosenIdentityOptionsData{
			App:            appData,
			IdentityOption: lpa.CertificateProviderIdentityOption,
		}

		return tmpl(w, data)
	}
}

func identityOptionPath(paths page.AppPaths, identityOption identity.Option) string {
	switch identityOption {
	case identity.OneLogin:
		return paths.CertificateProviderIdentityWithOneLogin
	case identity.EasyID:
		return paths.CertificateProviderIdentityWithYoti
	case identity.Passport:
		return paths.CertificateProviderIdentityWithPassport
	case identity.BiometricResidencePermit:
		return paths.CertificateProviderIdentityWithBiometricResidencePermit
	case identity.DrivingLicencePaper:
		return paths.CertificateProviderIdentityWithDrivingLicencePaper
	case identity.DrivingLicencePhotocard:
		return paths.CertificateProviderIdentityWithDrivingLicencePhotocard
	case identity.OnlineBankAccount:
		return paths.CertificateProviderIdentityWithOnlineBankAccount
	default:
		panic("missing case in identityOptionPath")
	}
}
