package supporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetViewLPA(t *testing.T) {
	testcases := map[string]error{
		"with actors":    nil,
		"without actors": dynamo.NotFoundError{},
	}

	for name, storeError := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

			donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

			ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(mock.Anything).
				Return("translated")
			localizer.EXPECT().
				Format(mock.Anything, mock.Anything).
				Return("translated")

			appDataWithLocalizer := page.AppData{Localizer: localizer}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Get(ctx).
				Return(donor, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				GetAny(ctx).
				Return(&actor.CertificateProviderProvidedDetails{}, storeError)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				GetAny(ctx).
				Return([]*actor.AttorneyProvidedDetails{{}}, storeError)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &viewLPAData{
					App:   appDataWithLocalizer,
					Donor: donor,
					Progress: actor.Progress{
						Paid:                      actor.ProgressTask{State: actor.TaskInProgress, Label: "translated"},
						ConfirmedID:               actor.ProgressTask{Label: "translated"},
						DonorSigned:               actor.ProgressTask{Label: "translated"},
						CertificateProviderSigned: actor.ProgressTask{Label: "translated"},
						AttorneysSigned:           actor.ProgressTask{Label: "translated"},
						LpaSubmitted:              actor.ProgressTask{Label: "translated"},
						StatutoryWaitingPeriod:    actor.ProgressTask{Label: "translated"},
						LpaRegistered:             actor.ProgressTask{Label: "translated"},
					},
				}).
				Return(nil)

			err := ViewLPA(template.Execute, donorStore, certificateProviderStore, attorneyStore)(appDataWithLocalizer, w, r, &actor.Organisation{})

			assert.Nil(t, err)
		})
	}

}

func TestGetViewLPAWithSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/", nil)

	err := ViewLPA(nil, nil, nil, nil)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAMissingLPAId(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ViewLPA(nil, nil, nil, nil)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAWithDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(nil, expectedError)

	err := ViewLPA(nil, donorStore, nil, nil)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAWhenCertificateProviderStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := ViewLPA(nil, donorStore, certificateProviderStore, nil)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAWhenAttorneyStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(mock.Anything).
		Return(nil, expectedError)

	err := ViewLPA(nil, donorStore, certificateProviderStore, attorneyStore)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id", SK: "ORGANISATION"}

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("translated")
	localizer.EXPECT().
		Format(mock.Anything, mock.Anything).
		Return("translated")

	appDataWithLocalizer := page.AppData{Localizer: localizer}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(mock.Anything).
		Return(donor, nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(mock.Anything).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(mock.Anything).
		Return([]*actor.AttorneyProvidedDetails{{}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ViewLPA(template.Execute, donorStore, certificateProviderStore, attorneyStore)(appDataWithLocalizer, w, r, &actor.Organisation{})

	assert.Error(t, err)
}
