package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseNewCertificateProviderData struct {
	Donor  *donordata.Provided
	Errors validation.List
	App    appcontext.Data
}

func ChooseNewCertificateProvider(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &chooseNewCertificateProviderData{Donor: provided, App: appData}

		if r.Method == http.MethodPost {
			provided.CertificateProvider = donordata.CertificateProvider{}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return donor.PathChooseYourCertificateProvider.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
