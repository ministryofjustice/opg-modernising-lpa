package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaProgress(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("translated")
	localizer.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return("translated")

	appDataWithLocalizer := page.AppData{Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaProgressData{
			App:   appDataWithLocalizer,
			Donor: &actor.DonorProvidedDetails{LpaID: "123"},
			Progress: actor.Progress{
				DonorSigned:               actor.ProgressTask{Label: "translated", State: actor.TaskInProgress},
				CertificateProviderSigned: actor.ProgressTask{Label: "translated"},
				AttorneysSigned:           actor.ProgressTask{Label: "translated"},
				LpaSubmitted:              actor.ProgressTask{Label: "translated"},
				StatutoryWaitingPeriod:    actor.ProgressTask{Label: "translated"},
				LpaRegistered:             actor.ProgressTask{Label: "translated"},
			},
		}).
		Return(nil)

	err := LpaProgress(template.Execute, certificateProviderStore, attorneyStore)(appDataWithLocalizer, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaProgressWhenCertificateProviderStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, expectedError)

	err := LpaProgress(nil, certificateProviderStore, attorneyStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}

func TestGetLpaProgressOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		GetAny(r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("translated")
	localizer.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return("translated")

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := LpaProgress(template.Execute, certificateProviderStore, attorneyStore)(page.AppData{Localizer: localizer}, w, r, &actor.DonorProvidedDetails{LpaID: "123"})
	assert.Equal(t, expectedError, err)
}
