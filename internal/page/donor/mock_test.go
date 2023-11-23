package donor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

var mockUuidString = func() string { return "123" }

var (
	testAddress = place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "E",
		Country:    "GB",
	}
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID:    "session-id",
		LpaID:        "lpa-id",
		Lang:         localize.En,
		Paths:        page.Paths,
		AppPublicURL: "http://example.org",
	}
	testNow   = time.Date(2023, time.July, 3, 4, 5, 6, 1, time.UTC)
	testNowFn = func() time.Time { return testNow }
)

func (m *mockDonorStore) willReturnEmptyLpa(r *http.Request) *mockDonorStore {
	m.
		On("Get", r.Context()).
		Return(&actor.DonorProvidedDetails{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
		}, nil)

	return m
}

func (m *mockDonorStore) withCompletedPaymentLpaData(r *http.Request, paymentId, paymentReference string, paymentAmount int) *mockDonorStore {
	m.
		On("Put", r.Context(), &actor.DonorProvidedDetails{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []actor.Payment{{
				PaymentId:        paymentId,
				PaymentReference: paymentReference,
				Amount:           paymentAmount,
			}},
			Tasks: actor.DonorTasks{
				PayForLpa: actor.PaymentTaskCompleted,
			},
		}).
		Return(nil)

	return m
}

func (m *mockSessionStore) withPaySession(r *http.Request) *mockSessionStore {
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

func (m *mockSessionStore) withExpiredPaySession(r *http.Request, w *httptest.ResponseRecorder) *mockSessionStore {
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
