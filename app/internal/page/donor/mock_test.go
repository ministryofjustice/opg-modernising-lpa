package donor

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/mock"
)

const formUrlEncoded = "application/x-www-form-urlencoded"

var (
	expectedError = errors.New("err")
	appData       = page.AppData{
		SessionID: "session-id",
		LpaID:     "lpa-id",
		Lang:      localize.En,
		Paths:     page.Paths,
	}
)

type mockLpaStore struct {
	mock.Mock
}

func (m *mockLpaStore) Create(ctx context.Context) (*page.Lpa, error) {
	args := m.Called(ctx)

	return args.Get(0).(*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) Get(ctx context.Context) (*page.Lpa, error) {
	args := m.Called(ctx)
	return args.Get(0).(*page.Lpa), args.Error(1)
}

func (m *mockLpaStore) Put(ctx context.Context, v *page.Lpa) error {
	return m.Called(ctx, v).Error(0)
}

type mockTemplate struct {
	mock.Mock
}

func (m *mockTemplate) Func(w io.Writer, data interface{}) error {
	args := m.Called(w, data)
	return args.Error(0)
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Print(v ...interface{}) {
	m.Called(v...)
}

type mockYotiClient struct {
	mock.Mock
}

func (m *mockYotiClient) IsTest() bool {
	return m.Called().Bool(0)
}

func (m *mockYotiClient) SdkID() string {
	return m.Called().String(0)
}

func (m *mockYotiClient) User(token string) (identity.UserData, error) {
	args := m.Called(token)

	return args.Get(0).(identity.UserData), args.Error(1)
}

type mockOneLoginClient struct {
	mock.Mock
}

func (m *mockOneLoginClient) AuthCodeURL(state, nonce, locale string, identity bool) string {
	args := m.Called(state, nonce, locale, identity)
	return args.String(0)
}

func (m *mockOneLoginClient) Exchange(ctx context.Context, code, nonce string) (string, error) {
	args := m.Called(ctx, code, nonce)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockOneLoginClient) UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error) {
	args := m.Called(ctx, accessToken)
	return args.Get(0).(onelogin.UserInfo), args.Error(1)
}

func (m *mockOneLoginClient) ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error) {
	args := m.Called(ctx, userInfo)
	return args.Get(0).(identity.UserData), args.Error(1)
}

type mockAddressClient struct {
	mock.Mock
}

func (m *mockAddressClient) LookupPostcode(ctx context.Context, postcode string) ([]place.Address, error) {
	args := m.Called(ctx, postcode)
	return args.Get(0).([]place.Address), args.Error(1)
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

type mockPayClient struct {
	mock.Mock
	BaseURL string
}

func (m *mockPayClient) CreatePayment(body pay.CreatePaymentBody) (pay.CreatePaymentResponse, error) {
	args := m.Called(body)
	return args.Get(0).(pay.CreatePaymentResponse), args.Error(1)
}

func (m *mockPayClient) GetPayment(paymentId string) (pay.GetPaymentResponse, error) {
	args := m.Called(paymentId)
	return args.Get(0).(pay.GetPaymentResponse), args.Error(1)
}

type mockNotifyClient struct {
	mock.Mock
}

func (m *mockNotifyClient) TemplateID(id notify.TemplateId) string {
	return m.Called(id).String(0)
}

func (m *mockNotifyClient) Email(ctx context.Context, email notify.Email) (string, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Error(1)
}

func (m *mockNotifyClient) Sms(ctx context.Context, sms notify.Sms) (string, error) {
	args := m.Called(ctx, sms)
	return args.String(0), args.Error(1)
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
