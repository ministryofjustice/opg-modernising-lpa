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

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &page.MockTemplate{}
	template.
		On("Func", w, &identityWithYotiData{
			App:         page.TestAppData,
			ClientSdkID: "an-sdk-id",
			ScenarioID:  "a-scenario-id",
		}).
		Return(nil)

	err := IdentityWithYoti(template.Func, lpaStore, yotiClient, "a-scenario-id")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient, template)
}

func TestGetIdentityWithYotiWhenAlreadyProvided(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{YotiUserData: identity.UserData{OK: true}}, nil)

	err := IdentityWithYoti(nil, lpaStore, nil, "")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithYotiWhenTest(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(true)

	err := IdentityWithYoti(nil, lpaStore, yotiClient, "")(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithYotiCallback, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient)
}

func TestGetIdentityWithYotiWhenDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, page.ExpectedError)

	err := IdentityWithYoti(nil, lpaStore, nil, "a-scenario-id")(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetIdentityWithYotiWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := &mockYotiClient{}
	yotiClient.On("IsTest").Return(false)
	yotiClient.On("SdkID").Return("an-sdk-id")

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(page.ExpectedError)

	err := IdentityWithYoti(template.Func, lpaStore, yotiClient, "a-scenario-id")(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, yotiClient, template)
}
