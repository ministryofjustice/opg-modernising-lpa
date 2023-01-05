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

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)
	lpaStore.On("Put", mock.Anything, "session-id", &Lpa{YotiUserData: userData}).Return(nil)

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

	err := IdentityWithYotiCallback(template.Func, yotiClient, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient, template)
}

func TestGetIdentityWithYotiCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, yotiClient, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient)
}

func TestGetIdentityWithYotiCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithYotiCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Now()
	userData := identity.UserData{FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{}, nil)
	lpaStore.On("Put", mock.Anything, "session-id", &Lpa{YotiUserData: userData}).Return(expectedError)

	yotiClient := &mockYotiClient{}
	yotiClient.On("User", "a-token").Return(userData, nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(nil, yotiClient, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient)
}

func TestGetIdentityWithYotiCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, FullName: "a-full-name", RetrievedAt: now}

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{YotiUserData: userData}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiCallbackData{
			App:         appData,
			FullName:    "a-full-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	err := IdentityWithYotiCallback(template.Func, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}

func TestPostIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.On("Get", mock.Anything, "session-id").Return(&Lpa{
		IdentityOption: EasyID,
	}, nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithYotiCallback(nil, nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, appData.Paths.ReadYourLpa, resp.Header.Get("Location"))
}
