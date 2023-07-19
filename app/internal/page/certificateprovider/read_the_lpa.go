package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type readTheLpaData struct {
	App                 page.AppData
	Errors              validation.List
	Lpa                 *page.Lpa
	CertificateProvider *actor.CertificateProviderProvidedDetails
}

func ReadTheLpa(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ReadTheLpa = actor.TaskCompleted
			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			if !lpa.Submitted.IsZero() {
				return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.ProvideCertificate.Format(lpa.ID))
			}

			return appData.Redirect(w, r, nil, page.Paths.CertificateProvider.WhatHappensNext.Format(lpa.ID))
		}

		data := &readTheLpaData{
			App:                 appData,
			Lpa:                 lpa,
			CertificateProvider: certificateProvider,
		}

		return tmpl(w, data)
	}
}
