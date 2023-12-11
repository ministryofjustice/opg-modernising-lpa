package page

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

var MockRandom = func(int) string { return "123" }

var (
	expectedError = errors.New("err")
	TestAppData   = AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     Paths,
	}
)

func mockDonorStoreWillReturnEmptyLpa(m *mockDonorStore, r *http.Request) *mockDonorStore {
	m.
		On("Get", r.Context()).
		Return(&actor.DonorProvidedDetails{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
		}, nil)

	return m
}

func mockDonorStoreWithCompletedPaymentLpaData(m *mockDonorStore, r *http.Request, paymentId, paymentReference string) *mockDonorStore {
	m.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []actor.Payment{{
				PaymentId:        paymentId,
				PaymentReference: paymentReference,
			}},
			Tasks: actor.DonorTasks{
				PayForLpa: actor.PaymentTaskCompleted,
			},
		}).
		Return(nil)

	return m
}

func (m *mockSessionStore) WithPaySession(r *http.Request) *mockSessionStore {
	getSession := sessions.NewSession(m, "pay")

	getSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   5400,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	getSession.Values = map[any]any{"payment": &sesh.PaymentSession{PaymentID: "abc123"}}

	m.On("Get", r, "pay").Return(getSession, nil)

	return m
}

func (m *mockSessionStore) WithExpiredPaySession(r *http.Request, w *httptest.ResponseRecorder) *mockSessionStore {
	storeSession := sessions.NewSession(m, "pay")

	// Expire cookie
	storeSession.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}
	storeSession.Values = map[any]any{}
	m.On("Save", r, w, storeSession).Return(nil)

	return m
}
