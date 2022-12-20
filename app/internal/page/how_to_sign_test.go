package page

//
//import (
//	"context"
//	"net/http"
//	"net/http/httptest"
//	"testing"
//
//	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/mock"
//)
//
//type mockNotifyClient struct {
//	mock.Mock
//}
//
//func (m *mockNotifyClient) TemplateID(name string) string {
//	return m.Called(name).String(0)
//}
//
//func (m *mockNotifyClient) Email(ctx context.Context, email notify.Email) (string, error) {
//	args := m.Called(ctx, email)
//	return args.String(0), args.Error(1)
//}
//
//func TestGetHowToSign(t *testing.T) {
//	w := httptest.NewRecorder()
//	lpa := &Lpa{You: Person{Email: "me@example.com"}}
//
//	lpaStore := &mockLpaStore{}
//	lpaStore.
//		On("Get", mock.Anything, "session-id").
//		Return(lpa, nil)
//
//	template := &mockTemplate{}
//	template.
//		On("Func", w, &howToSignData{App: appData, Lpa: lpa}).
//		Return(nil)
//
//	r, _ := http.NewRequest(http.MethodGet, "/", nil)
//
//	err := HowToSign(template.Func, lpaStore, nil, nil)(appData, w, r)
//	resp := w.Result()
//
//	assert.Nil(t, err)
//	assert.Equal(t, http.StatusOK, resp.StatusCode)
//	mock.AssertExpectationsForObjects(t, lpaStore, template)
//}
//
//func TestGetHowToSignWhenLpaStoreErrors(t *testing.T) {
//	w := httptest.NewRecorder()
//
//	lpaStore := &mockLpaStore{}
//	lpaStore.
//		On("Get", mock.Anything, "session-id").
//		Return(&Lpa{}, expectedError)
//
//	r, _ := http.NewRequest(http.MethodGet, "/", nil)
//
//	err := HowToSign(nil, lpaStore, nil, nil)(appData, w, r)
//
//	assert.Equal(t, expectedError, err)
//	mock.AssertExpectationsForObjects(t, lpaStore)
//}
//
//func TestGetHowToSignWhenTemplateErrors(t *testing.T) {
//	w := httptest.NewRecorder()
//	lpa := &Lpa{You: Person{Email: "me@example.com"}}
//
//	lpaStore := &mockLpaStore{}
//	lpaStore.
//		On("Get", mock.Anything, "session-id").
//		Return(lpa, nil)
//
//	template := &mockTemplate{}
//	template.
//		On("Func", w, &howToSignData{App: appData, Lpa: lpa}).
//		Return(expectedError)
//
//	r, _ := http.NewRequest(http.MethodGet, "/", nil)
//
//	err := HowToSign(template.Func, lpaStore, nil, nil)(appData, w, r)
//
//	assert.Equal(t, expectedError, err)
//	mock.AssertExpectationsForObjects(t, lpaStore, template)
//}
//
//func TestPostHowToSign(t *testing.T) {
//	w := httptest.NewRecorder()
//	lpa := &Lpa{You: Person{Email: "me@example.com"}}
//
//	lpaStore := &mockLpaStore{}
//	lpaStore.
//		On("Get", mock.Anything, "session-id").
//		Return(lpa, nil)
//	lpaStore.
//		On("Put", mock.Anything, "session-id", &Lpa{
//			You:              Person{Email: "me@example.com"},
//			SignatureCode:    "1234",
//			SignatureEmailID: "email-id",
//		}).
//		Return(nil)
//
//	notifyClient := &mockNotifyClient{}
//	notifyClient.
//		On("TemplateID", "MLPA Beta signature code").
//		Return("xyz")
//	notifyClient.
//		On("Email", mock.Anything, notify.Email{
//			EmailAddress:    "me@example.com",
//			TemplateID:      "xyz",
//			Personalisation: map[string]string{"code": "1234"},
//		}).
//		Return("email-id", nil)
//
//	r, _ := http.NewRequest(http.MethodPost, "/", nil)
//
//	err := HowToSign(nil, lpaStore, notifyClient, func(l int) string { return "1234" })(appData, w, r)
//	resp := w.Result()
//
//	assert.Nil(t, err)
//	assert.Equal(t, http.StatusFound, resp.StatusCode)
//	assert.Equal(t, appData.Paths.ReadYourLpa, resp.Header.Get("Location"))
//	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
//}
//
//func TestPostHowToSignWhenNotifyErrors(t *testing.T) {
//	w := httptest.NewRecorder()
//	lpa := &Lpa{You: Person{Email: "me@example.com"}}
//
//	lpaStore := &mockLpaStore{}
//	lpaStore.
//		On("Get", mock.Anything, "session-id").
//		Return(lpa, nil)
//
//	notifyClient := &mockNotifyClient{}
//	notifyClient.
//		On("TemplateID", "MLPA Beta signature code").
//		Return("xyz")
//	notifyClient.
//		On("Email", mock.Anything, mock.Anything).
//		Return("", expectedError)
//
//	r, _ := http.NewRequest(http.MethodPost, "/", nil)
//
//	err := HowToSign(nil, lpaStore, notifyClient, func(l int) string { return "1234" })(appData, w, r)
//
//	assert.Equal(t, expectedError, err)
//	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
//}
//
//func TestPostHowToSignWhenLpaStoreErrors(t *testing.T) {
//	w := httptest.NewRecorder()
//	lpa := &Lpa{You: Person{Email: "me@example.com"}}
//
//	lpaStore := &mockLpaStore{}
//	lpaStore.
//		On("Get", mock.Anything, "session-id").
//		Return(lpa, nil)
//	lpaStore.
//		On("Put", mock.Anything, "session-id", mock.Anything).
//		Return(expectedError)
//
//	notifyClient := &mockNotifyClient{}
//	notifyClient.
//		On("TemplateID", "MLPA Beta signature code").
//		Return("xyz")
//	notifyClient.
//		On("Email", mock.Anything, mock.Anything).
//		Return("email-id", nil)
//
//	r, _ := http.NewRequest(http.MethodPost, "/", nil)
//
//	err := HowToSign(nil, lpaStore, notifyClient, func(l int) string { return "1234" })(appData, w, r)
//
//	assert.Equal(t, expectedError, err)
//	mock.AssertExpectationsForObjects(t, lpaStore, notifyClient)
//}
