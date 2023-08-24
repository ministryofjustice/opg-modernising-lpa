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

func YourChosenIdentityOptions(tmpl template.Template, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, nil, identityOptionPath(appData.Paths, certificateProvider.IdentityOption).Format(certificateProvider.LpaID))
		}

		data := &yourChosenIdentityOptionsData{
			App:            appData,
			IdentityOption: certificateProvider.IdentityOption,
		}

		return tmpl(w, data)
	}
}

func identityOptionPath(paths page.AppPaths, identityOption identity.Option) interface{ Format(string) string } {
	switch identityOption {
	case identity.OneLogin:
		return paths.CertificateProvider.IdentityWithOneLogin
	case identity.EasyID:
		return paths.CertificateProvider.IdentityWithYoti
	case identity.Passport:
		return paths.CertificateProvider.IdentityWithPassport
	case identity.BiometricResidencePermit:
		return paths.CertificateProvider.IdentityWithBiometricResidencePermit
	case identity.DrivingLicencePaper:
		return paths.CertificateProvider.IdentityWithDrivingLicencePaper
	case identity.DrivingLicencePhotocard:
		return paths.CertificateProvider.IdentityWithDrivingLicencePhotocard
	case identity.OnlineBankAccount:
		return paths.CertificateProvider.IdentityWithOnlineBankAccount
	default:
		panic("missing case in identityOptionPath")
	}
}
