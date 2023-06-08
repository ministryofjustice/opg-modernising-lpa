package page

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dashboardData struct {
	App                     AppData
	Errors                  validation.List
	UseTabs                 bool
	DonorLpas               []*Lpa
	CertificateProviderLpas []*Lpa
	AttorneyLpas            []*Lpa
}

func Dashboard(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			lpa, err := donorStore.Create(r.Context())
			if err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, Paths.YourDetails)
		}

		donorLpas, err := donorStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		certificateProviderDetails, err := certificateProviderStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		certificateProviderLpas := make([]*Lpa, len(certificateProviderDetails))
		for i, detail := range certificateProviderDetails {
			lpa, err := donorStore.GetAny(ContextWithSessionData(context.Background(), &SessionData{LpaID: detail.LpaID}))
			if err != nil {
				return err
			}

			certificateProviderLpas[i] = lpa
		}

		attorneyDetails, err := attorneyStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		attorneyLpas := make([]*Lpa, len(attorneyDetails))
		for i, detail := range attorneyDetails {
			lpa, err := donorStore.GetAny(ContextWithSessionData(context.Background(), &SessionData{LpaID: detail.LpaID}))
			if err != nil {
				return err
			}

			attorneyLpas[i] = lpa
		}

		tabCount := 0
		if len(donorLpas) > 0 {
			tabCount++
		}
		if len(certificateProviderLpas) > 0 {
			tabCount++
		}
		if len(attorneyLpas) > 0 {
			tabCount++
		}

		data := &dashboardData{
			App:                     appData,
			UseTabs:                 tabCount > 1,
			DonorLpas:               donorLpas,
			CertificateProviderLpas: certificateProviderLpas,
			AttorneyLpas:            attorneyLpas,
		}

		return tmpl(w, data)
	}
}
