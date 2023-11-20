package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readTheLpaData struct {
	App                 page.AppData
	Errors              validation.List
	Donor               *actor.DonorProvidedDetails
	CertificateProvider *actor.CertificateProviderProvidedDetails
}

func ReadTheLpa(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		donor, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if donor.SignedAt.IsZero() || !donor.Tasks.PayForLpa.IsCompleted() {
				return page.Paths.CertificateProvider.TaskList.Redirect(w, r, appData, donor.LpaID)
			}

			certificateProvider.Tasks.ReadTheLpa = actor.TaskCompleted
			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.WhatHappensNext.Redirect(w, r, appData, donor.LpaID)
		}

		data := &readTheLpaData{
			App:                 appData,
			Donor:               donor,
			CertificateProvider: certificateProvider,
		}

		return tmpl(w, data)
	}
}
