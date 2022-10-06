package page

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	userData := identity.UserData{FullName: "a-full-name", RetrievedAt: now}

	dataStore := &mockDataStore{}
	dataStore.On("Get", mock.Anything, "session-id").Return(nil)
	dataStore.On("Put", mock.Anything, "session-id", Lpa{YotiUserData: userData}).Return(nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(userData, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(template.Func, yotiClient, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, yotiClient, template)
}

func TestGetIdentityWithYotiCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.On("Get", mock.Anything, "session-id").Return(nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, yotiClient, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore, yotiClient)
}

func TestGetIdentityWithYotiCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.On("Get", mock.Anything, "session-id").Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetIdentityWithYotiCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	userData := identity.UserData{FullName: "a-full-name", RetrievedAt: now}

	dataStore := &mockDataStore{}
	dataStore.On("Get", mock.Anything, "session-id").Return(nil)
	dataStore.On("Put", mock.Anything, "session-id", Lpa{YotiUserData: userData}).Return(expectedError)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(userData, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, yotiClient, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore, yotiClient)
}

func TestGetIdentityWithYotiCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	dataStore := &mockDataStore{data: Lpa{YotiUserData: userData}}
	dataStore.On("Get", mock.Anything, "session-id").Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(template.Func, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, template)
}

func TestPostIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{data: Lpa{IdentityOptions: IdentityOptions{First: Yoti, Second: Passport}}}
	dataStore.On("Get", mock.Anything, "session-id").Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithYotiCallback(nil, nil, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityWithPassportPath, resp.Header.Get("Location"))
}
