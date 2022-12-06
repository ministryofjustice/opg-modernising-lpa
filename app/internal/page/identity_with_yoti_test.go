package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
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

func TestGetIdentityWithYoti(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiData{
			App:         appData,
			ClientSdkID: "an-sdk-id",
			ScenarioID:  "a-scenario-id",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(template.Func, lpaStore, yotiClient, "a-scenario-id")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient, template)
}

func TestGetIdentityWithYotiWhenAlreadyProvided(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{YotiUserData: identity.UserData{OK: true}}, nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(nil, lpaStore, nil, "")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithYotiWhenTest(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(true)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(nil, lpaStore, yotiClient, "")(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient)
}

func TestGetIdentityWithYotiWhenDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(nil, lpaStore, nil, "a-scenario-id")(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithYotiWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := IdentityWithYoti(template.Func, lpaStore, yotiClient, "a-scenario-id")(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient, template)
}
