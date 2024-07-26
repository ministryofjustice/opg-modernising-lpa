package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readTheLpaData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *lpastore.Lpa
}

func ReadTheLpa(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if lpa.SignedAt.IsZero() || !lpa.Paid {
				return page.Paths.CertificateProvider.TaskList.Redirect(w, r, appData, lpa.LpaID)
			}

			certificateProvider.Tasks.ReadTheLpa = actor.TaskCompleted
			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			return page.Paths.CertificateProvider.WhatHappensNext.Redirect(w, r, appData, lpa.LpaID)
		}

		data := &readTheLpaData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
