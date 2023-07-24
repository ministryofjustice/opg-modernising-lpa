package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type taskListData struct {
	App    page.AppData
	Errors validation.List
	Lpa    *page.Lpa
	Items  []taskListItem
}

type taskListItem struct {
	Name     string
	Path     string
	State    actor.TaskState
	Disabled bool
}

func TaskList(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		tasks := certificateProvider.Tasks

		data := &taskListData{
			App: appData,
			Lpa: lpa,
			Items: []taskListItem{
				{
					Name:  "confirmYourDetails",
					Path:  page.Paths.CertificateProvider.EnterDateOfBirth.Format(lpa.ID),
					State: tasks.ConfirmYourDetails,
				},
				{
					Name:     "confirmYourIdentity",
					Path:     page.Paths.CertificateProvider.WhatYoullNeedToConfirmYourIdentity.Format(lpa.ID),
					State:    tasks.ConfirmYourIdentity,
					Disabled: !lpa.Tasks.PayForLpa.IsCompleted() || lpa.Submitted.IsZero(),
				},
				{
					Name:     "readTheLpa",
					Path:     page.Paths.CertificateProvider.ReadTheLpa.Format(lpa.ID),
					State:    tasks.ReadTheLpa,
					Disabled: lpa.Submitted.IsZero(),
				},
				{
					Name:     "provideTheCertificateForThisLpa",
					Path:     page.Paths.CertificateProvider.ProvideCertificate.Format(lpa.ID),
					State:    tasks.ProvideTheCertificate,
					Disabled: lpa.Submitted.IsZero() || !tasks.ConfirmYourDetails.Completed() || !tasks.ConfirmYourIdentity.Completed() || !tasks.ReadTheLpa.Completed(),
				},
			},
		}

		return tmpl(w, data)
	}
}
