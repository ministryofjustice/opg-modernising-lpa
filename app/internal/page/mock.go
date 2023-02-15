package page

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/mock"
)

var MockRandom = func(int) string { return "123" }

var (
	TestAddress = place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "e",
	}
	ExpectedError = errors.New("err")
	TestAppData   = AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     Paths,
	}
)

type MockLpaStore struct {
	mock.Mock
}

func (m *MockLpaStore) Create(ctx context.Context) (*Lpa, error) {
	args := m.Called(ctx)

	return args.Get(0).(*Lpa), args.Error(1)
}

func (m *MockLpaStore) GetAll(ctx context.Context) ([]*Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*Lpa), args.Error(1)
}

func (m *MockLpaStore) Get(ctx context.Context) (*Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).(*Lpa), args.Error(1)
}

func (m *MockLpaStore) Put(ctx context.Context, v *Lpa) error {
	return m.Called(ctx, v).Error(0)
}

func (m *MockLpaStore) WillReturnEmptyLpa(r *http.Request) *MockLpaStore {
	m.
		On("Get", r.Context()).
		Return(&Lpa{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
		}, nil)

	return m
}

func (m *MockLpaStore) WithCompletedPaymentLpaData(r *http.Request, paymentId, paymentReference string) *MockLpaStore {
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

type MockTemplate struct {
	mock.Mock
}

func (m *MockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Print(v ...interface{}) {
	m.Called(v...)
}

type MockOneLoginClient struct {
	mock.Mock
}

func (m *MockOneLoginClient) AuthCodeURL(state, nonce, locale string, identity bool) string {
	args := m.Called(state, nonce, locale, identity)
	return args.String(0)
}

func (m *MockOneLoginClient) Exchange(ctx context.Context, code, nonce string) (string, error) {
	args := m.Called(ctx, code, nonce)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockOneLoginClient) UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(onelogin.UserInfo), args.Error(1)
}

func (m *MockOneLoginClient) ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error) {
	args := m.Called(ctx, userInfo)
	return args.Get(0).(identity.UserData), args.Error(1)
}

type MockAddressClient struct {
	mock.Mock
}

func (m *MockAddressClient) LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error) {
	args := m.Called(ctx, postcode)
	return args.Get(0).([]place.Address), args.Error(1)
}

type MockSessionsStore struct {
	mock.Mock
}

func (m *MockSessionsStore) New(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *MockSessionsStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	args := m.Called(r, name)
	return args.Get(0).(*sessions.Session), args.Error(1)
}

func (m *MockSessionsStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	args := m.Called(r, w, session)
	return args.Error(0)
}

func (m *MockSessionsStore) WithPaySession(r *http.Request) *MockSessionsStore {
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

func (m *MockSessionsStore) WithExpiredPaySession(r *http.Request, w *httptest.ResponseRecorder) *MockSessionsStore {
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
