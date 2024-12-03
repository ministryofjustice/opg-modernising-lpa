package page

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DashboardStore interface {
	GetAll(ctx context.Context) (results dashboarddata.Results, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type dashboardForm struct {
	hasExistingDonorLPAs bool
}

type dashboardData struct {
	App                     appcontext.Data
	Errors                  validation.List
	NeedsTabs               bool
	DonorLpas               []dashboarddata.Actor
	RegisteredDonorLpas     []dashboarddata.Actor
	CertificateProviderLpas []dashboarddata.Actor
	AttorneyLpas            []dashboarddata.Actor
	RegisteredAttorneyLpas  []dashboarddata.Actor
	VoucherLpas             []dashboarddata.Actor
	UseURL                  string
}

func Dashboard(tmpl template.Template, donorStore DonorStore, dashboardStore DashboardStore, useURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		if r.Method == http.MethodPost {
			form := readDashboardForm(r)

			lpa, err := donorStore.Create(r.Context())
			if err != nil {
				return err
			}

			path := donor.PathYourName
			if form.hasExistingDonorLPAs {
				path = donor.PathMakeANewLPA
			}

			return path.Redirect(w, r, appData, lpa)
		}

		results, err := dashboardStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		var donorLpas, registeredDonorLpas []dashboarddata.Actor
		for _, lpa := range results.Donor {
			if lpa.Lpa.RegisteredAt.IsZero() {
				donorLpas = append(donorLpas, lpa)
			} else {
				registeredDonorLpas = append(registeredDonorLpas, lpa)
			}
		}

		var attorneyLpas, registeredAttorneyLpas []dashboarddata.Actor
		for _, lpa := range results.Attorney {
			if lpa.Lpa.RegisteredAt.IsZero() {
				attorneyLpas = append(attorneyLpas, lpa)
			} else {
				registeredAttorneyLpas = append(registeredAttorneyLpas, lpa)
			}
		}

		data := &dashboardData{
			App:                     appData,
			NeedsTabs:               len(results.CertificateProvider) > 0 || len(results.Attorney) > 0 || len(results.Voucher) > 0,
			DonorLpas:               donorLpas,
			RegisteredDonorLpas:     registeredDonorLpas,
			CertificateProviderLpas: results.CertificateProvider,
			AttorneyLpas:            attorneyLpas,
			RegisteredAttorneyLpas:  registeredAttorneyLpas,
			VoucherLpas:             results.Voucher,
			UseURL:                  useURL,
		}

		return tmpl(w, data)
	}
}

func readDashboardForm(r *http.Request) *dashboardForm {
	f := &dashboardForm{}
	f.hasExistingDonorLPAs = r.PostFormValue("has-existing-donor-lpas") == "true"
	return f
}
