package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Now()
	userData := identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor: actor.Donor{FirstNames: "first-name", LastName: "last-name"},
		}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{
			Donor:            actor.Donor{FirstNames: "first-name", LastName: "last-name"},
			IdentityUserData: userData,
		}).
		Return(nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("User", "a-token").Return(userData, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithYotiCallbackData{
			App:         testAppData,
			FullName:    "first-name last-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithYotiCallback(template.Execute, yotiClient, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithYotiCallbackWhenError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("User", "a-token").Return(identity.UserData{}, expectedError)

	err := IdentityWithYotiCallback(nil, yotiClient, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithYotiCallbackWhenGetDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{}, expectedError)

	err := IdentityWithYotiCallback(nil, nil, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithYotiCallbackWhenPutDataStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Now()
	userData := identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor: actor.Donor{FirstNames: "first-name", LastName: "last-name"},
		}, nil)
	lpaStore.On("Put", r.Context(), mock.Anything).Return(expectedError)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("User", "a-token").Return(userData, nil)

	err := IdentityWithYotiCallback(nil, yotiClient, lpaStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithYotiCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor:            actor.Donor{FirstNames: "first-name", LastName: "last-name"},
			IdentityUserData: userData,
		}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithYotiCallbackData{
			App:         testAppData,
			FullName:    "first-name last-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithYotiCallback(template.Execute, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	userData := identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	lpaStore := newMockLpaStore(t)
	lpaStore.On("Get", r.Context()).Return(&page.Lpa{
		Donor:            actor.Donor{FirstNames: "first-name", LastName: "last-name"},
		IdentityUserData: userData,
		IdentityOption:   identity.EasyID,
	}, nil)

	err := IdentityWithYotiCallback(nil, nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}
