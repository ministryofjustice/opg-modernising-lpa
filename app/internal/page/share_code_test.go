package page

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestShareCodeSenderSend(t *testing.T) {
	ctx := context.Background()

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
		Return(nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", notify.TemplateId(99)).
		Return("template-id")
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "name@example.com",
			Personalisation: map[string]string{
				"link": fmt.Sprintf("http://app%s?share-code=123", Paths.CertificateProviderStart),
			},
		}).
		Return("", nil)

	sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, "name@example.com")

	assert.Nil(t, err)
	mock.AssertExpectationsForObjects(t, notifyClient, dataStore)
}

func TestShareCodeSenderSendWhenEmailErrors(t *testing.T) {
	ctx := context.Background()

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
		Return(nil)

	notifyClient := &mockNotifyClient{}
	notifyClient.
		On("TemplateID", notify.TemplateId(99)).
		Return("template-id")
	notifyClient.
		On("Email", ctx, notify.Email{
			TemplateID:   "template-id",
			EmailAddress: "name@example.com",
			Personalisation: map[string]string{
				"link": fmt.Sprintf("http://app%s?share-code=123", Paths.CertificateProviderStart),
			},
		}).
		Return("", ExpectedError)

	sender := NewShareCodeSender(dataStore, notifyClient, "http://app", MockRandom)
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, "name@example.com")

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
	mock.AssertExpectationsForObjects(t, notifyClient, dataStore)
}

func TestShareCodeSenderSendWhenDataStoreErrors(t *testing.T) {
	ctx := context.Background()

	dataStore := &mockDataStore{}
	dataStore.
		On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "session-id", LpaID: "lpa-id"}).
		Return(ExpectedError)

	sender := NewShareCodeSender(dataStore, nil, "http://app", MockRandom)
	err := sender.Send(ctx, notify.TemplateId(99), TestAppData, "name@example.com")

	assert.Equal(t, ExpectedError, errors.Unwrap(err))
	mock.AssertExpectationsForObjects(t, dataStore)
}
