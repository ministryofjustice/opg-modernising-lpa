package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"

	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCertificateProviderYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{
		ID: "lpa-id",
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "lpa-id",
			},
		}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &cpYourDetailsData{
			App:  appData,
			Lpa:  lpa,
			Form: &cpYourDetailsForm{},
		}).
		Return(nil)

	err := certificateProviderYourDetails(template.Func, lpaStore, sessionStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &Lpa{
		ID: "lpa-id",
		CertificateProviderProvidedDetails: actor.CertificateProvider{
			Email:       "a@example.org",
			Mobile:      "07535111222",
			DateOfBirth: date.New("1997", "1", "2"),
		},
	}
	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(lpa, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "lpa-id",
			},
		}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &cpYourDetailsData{
			App: appData,
			Lpa: lpa,
			Form: &cpYourDetailsForm{
				Email:  "a@example.org",
				Mobile: "07535111222",
				Dob:    date.New("1997", "1", "2"),
			},
		}).
		Return(nil)

	err := certificateProviderYourDetails(template.Func, lpaStore, sessionStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := certificateProviderYourDetails(nil, lpaStore, nil)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetCertificateProviderYourDetailsWhenSessionStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{}, expectedError)

	err := certificateProviderYourDetails(nil, lpaStore, sessionStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
}

func TestGetCertificateProviderYourDetailsWhenLpaIdMismatch(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{ID: "lpa-id"}, nil)

	sessionStore := &mockSessionsStore{}
	sessionStore.
		On("Get", r, "session").
		Return(&sessions.Session{Values: map[any]any{
			"certificate-provider": &CertificateProviderSession{
				Sub:            "random",
				DonorSessionID: "session-id",
				LpaID:          "not-lpa-id",
			},
		}}, nil)

	err := certificateProviderYourDetails(nil, lpaStore, sessionStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("/lpa/lpa-id%s", appData.Paths.CertificateProviderStart), resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, sessionStore)
}
