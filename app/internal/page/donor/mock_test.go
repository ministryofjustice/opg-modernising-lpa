package donor

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/mock"
)

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
