package page

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/mock"
)

var MockRandom = func(int) string { return "123" }

var (
	ExpectedError = errors.New("err")
	TestAppData   = AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     Paths,
	}
)

func mockLpaStoreWillReturnEmptyLpa(m *mockLpaStore, r *http.Request) *mockLpaStore {
	m.
		On("Get", r.Context()).
		Return(&Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
		}, nil)

	return m
}

func mockLpaStoreWithCompletedPaymentLpaData(m *mockLpaStore, r *http.Request, paymentId, paymentReference string) *mockLpaStore {
	m.
		On("Put", r.Context(), &Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: PaymentDetails{
				PaymentId:        paymentId,
				PaymentReference: paymentReference,
			},
			Tasks: Tasks{
				PayForLpa: TaskCompleted,
			},
		}).
		Return(nil)

	return m
}

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockSessionsStore struct {
	mock.Mock
}

func (m *mockSessionsStore) New(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *mockSessionsStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	args := m.Called(r, w, session)
	return args.Error(0)
}

func (m *mockSessionsStore) WithPaySession(r *http.Request) *mockSessionsStore {
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

func (m *mockSessionsStore) WithExpiredPaySession(r *http.Request, w *httptest.ResponseRecorder) *mockSessionsStore {
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
