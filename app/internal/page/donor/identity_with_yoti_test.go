package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithYoti(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &mockTemplate{}
	template.
		On("Func", w, &identityWithYotiData{
			App:         testAppData,
			ClientSdkID: "an-sdk-id",
			ScenarioID:  "a-scenario-id",
		}).
		Return(nil)

	err := IdentityWithYoti(template.Func, lpaStore, yotiClient, "a-scenario-id")(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetIdentityWithYotiWhenAlreadyProvided(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{YotiUserData: identity.UserData{OK: true}}, nil)

	err := IdentityWithYoti(nil, lpaStore, nil, "")(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
}

func TestGetIdentityWithYotiWhenTest(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("IsTest").Return(true)

	err := IdentityWithYoti(nil, lpaStore, yotiClient, "")(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
}

func TestGetIdentityWithYotiWhenDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, expectedError)

	err := IdentityWithYoti(nil, lpaStore, nil, "a-scenario-id")(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithYotiWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	err := IdentityWithYoti(template.Func, lpaStore, yotiClient, "a-scenario-id")(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, template)
}
