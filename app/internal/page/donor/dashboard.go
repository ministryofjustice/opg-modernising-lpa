package donor

import (
	"context"
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DashboardLpaDatum struct {
	Lpa                 *page.Lpa
	CertificateProvider *actor.CertificateProvider
}

type dashboardData struct {
	App    page.AppData
	Errors validation.List
	Lpas   []DashboardLpaDatum
}

func Dashboard(tmpl template.Template, lpaStore LpaStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			lpa, err := lpaStore.Create(r.Context())
			if err != nil {
				return err
			}

			return appData.Redirect(w, r, lpa, page.Paths.YourDetails)
		}

		lpas, err := lpaStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		dashboardLpaData, err := buildDashboardLpaData(lpas, certificateProviderStore, r.Context())
		if err != nil {
			return err
		}

		data := &dashboardData{
			App:  appData,
			Lpas: dashboardLpaData,
		}

		return tmpl(w, data)
	}
}

func buildDashboardLpaData(lpas []*page.Lpa, store page.CertificateProviderStore, ctx context.Context) ([]DashboardLpaDatum, error) {
	var dashboardLpaData []DashboardLpaDatum

	for _, lpa := range lpas {
		datum := DashboardLpaDatum{
			Lpa: lpa,
		}

		ctx := page.ContextWithSessionData(ctx, &page.SessionData{LpaID: lpa.ID})

		cp, err := store.Get(ctx)

		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return dashboardLpaData, err
		}

		if cp == nil {
			cp = &actor.CertificateProvider{}
		}

		datum.CertificateProvider = cp

		dashboardLpaData = append(dashboardLpaData, datum)
	}

	return dashboardLpaData, nil
}
