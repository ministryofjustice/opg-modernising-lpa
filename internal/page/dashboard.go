package page

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DashboardStore interface {
	GetAll(ctx context.Context) (donor, attorney, certificateProvider []LpaAndActorTasks, err error)
	SubExistsForActorType(ctx context.Context, sub string, actorType actor.Type) (bool, error)
}

type LpaAndActorTasks struct {
	Lpa                 *lpadata.Lpa
	CertificateProvider *certificateproviderdata.Provided
	Attorney            *attorneydata.Provided
}

type dashboardForm struct {
	hasExistingDonorLPAs bool
}

type dashboardData struct {
	App                     appcontext.Data
	Errors                  validation.List
	UseTabs                 bool
	DonorLpas               []LpaAndActorTasks
	CertificateProviderLpas []LpaAndActorTasks
	AttorneyLpas            []LpaAndActorTasks
}

func Dashboard(tmpl template.Template, donorStore DonorStore, dashboardStore DashboardStore) Handler {
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

		donorLpas, attorneyLpas, certificateProviderLpas, err := dashboardStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		tabCount := 1
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

func readDashboardForm(r *http.Request) *dashboardForm {
	f := &dashboardForm{}
	f.hasExistingDonorLPAs = r.PostFormValue("has-existing-donor-lpas") == "true"
	return f
}
