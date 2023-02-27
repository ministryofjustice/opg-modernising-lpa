package donor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"

	"github.com/stretchr/testify/mock"
)

var mockRandom = func(int) string { return "123" }

var (
	testAddress = place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
	}
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     page.Paths,
	}
)

func (m *mockLpaStore) willReturnEmptyLpa(r *http.Request) *mockLpaStore {
	m.
		On("Get", r.Context()).
		Return(&page.Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
		}, nil)

	return m
}

func (m *mockLpaStore) withCompletedPaymentLpaData(r *http.Request, paymentId, paymentReference string) *mockLpaStore {
	m.
		On("Put", r.Context(), &page.Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: page.PaymentDetails{
				PaymentId:        paymentId,
				PaymentReference: paymentReference,
			},
			Tasks: page.Tasks{
				PayForLpa: page.TaskCompleted,
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

type mockDataStore struct {
	data interface{}
	mock.Mock
}

func (m *mockDataStore) GetAll(ctx context.Context, pk string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, pk).Error(0)
}

func (m *mockDataStore) Get(ctx context.Context, pk, sk string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, pk, sk).Error(0)
}

func (m *mockDataStore) Put(ctx context.Context, pk, sk string, v interface{}) error {
	return m.Called(ctx, pk, sk, v).Error(0)
}
