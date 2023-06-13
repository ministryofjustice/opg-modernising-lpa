package donor

import (
	"io"
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

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Donor:                 actor.Donor{FirstNames: "first-name", LastName: "last-name"},
			DonorIdentityUserData: userData,
		}).
		Return(nil)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("User", "a-token").Return(userData, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithYotiCallbackData{
			App:         testAppData,
			FirstNames:  "first-name",
			LastName:    "last-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithYotiCallback(template.Execute, yotiClient, donorStore)(testAppData, w, r, &page.Lpa{
		Donor: actor.Donor{FirstNames: "first-name", LastName: "last-name"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithYotiCallbackWhenIdentityNotConfirmed(t *testing.T) {
	templateCalled := func(t *testing.T, w io.Writer) *mockTemplate {
		template := newMockTemplate(t)
		template.
			On("Execute", w, &identityWithYotiCallbackData{
				App:             testAppData,
				CouldNotConfirm: true,
			}).
			Return(nil)
		return template
	}

	templateIgnored := func(t *testing.T, w io.Writer) *mockTemplate {
		return nil
	}

	testCases := map[string]struct {
		yotiClient func(t *testing.T) *mockYotiClient
		template   func(*testing.T, io.Writer) *mockTemplate
		url        string
		error      error
	}{
		"not a match": {
			url: "/?code=a-code",
			yotiClient: func(t *testing.T) *mockYotiClient {
				yotiClient := newMockYotiClient(t)
				yotiClient.
					On("User", mock.Anything).
					Return(identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "x", LastName: "y"}, nil)
				return yotiClient
			},
			template: templateCalled,
		},
		"not ok": {
			url: "/?code=a-code",
			yotiClient: func(t *testing.T) *mockYotiClient {
				yotiClient := newMockYotiClient(t)
				yotiClient.
					On("User", mock.Anything).
					Return(identity.UserData{}, nil)
				return yotiClient
			},
			template: templateCalled,
		},
		"errored on user": {
			url: "/?code=a-code",
			yotiClient: func(t *testing.T) *mockYotiClient {
				yotiClient := newMockYotiClient(t)
				yotiClient.
					On("User", mock.Anything).
					Return(identity.UserData{OK: true}, expectedError)
				return yotiClient
			},
			template: templateIgnored,
			error:    expectedError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			yotiClient := tc.yotiClient(t)
			template := tc.template(t, w)

			err := IdentityWithYotiCallback(template.Execute, yotiClient, nil)(testAppData, w, r, &page.Lpa{})
			resp := w.Result()

			assert.Equal(t, tc.error, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetIdentityWithYotiCallbackWhenPutDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Now()
	userData := identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	donorStore := newMockDonorStore(t)
	donorStore.On("Put", r.Context(), mock.Anything).Return(expectedError)

	yotiClient := newMockYotiClient(t)
	yotiClient.On("User", "a-token").Return(userData, nil)

	err := IdentityWithYotiCallback(nil, yotiClient, donorStore)(testAppData, w, r, &page.Lpa{
		Donor: actor.Donor{FirstNames: "first-name", LastName: "last-name"},
	})

	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithYotiCallbackWhenReturning(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?token=a-token", nil)
	now := time.Date(2012, time.January, 1, 2, 3, 4, 5, time.UTC)
	userData := identity.UserData{OK: true, Provider: identity.EasyID, FirstNames: "first-name", LastName: "last-name", RetrievedAt: now}

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithYotiCallbackData{
			App:         testAppData,
			FirstNames:  "first-name",
			LastName:    "last-name",
			ConfirmedAt: now,
		}).
		Return(nil)

	err := IdentityWithYotiCallback(template.Execute, nil, nil)(testAppData, w, r, &page.Lpa{
		Donor:                 actor.Donor{FirstNames: "first-name", LastName: "last-name"},
		DonorIdentityUserData: userData,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostIdentityWithYotiCallback(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithYotiCallback(nil, nil, nil)(testAppData, w, r, &page.Lpa{
		DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.EasyID},
		DonorIdentityOption:   identity.EasyID,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}

func TestPostIdentityWithYotiCallbackNotConfirmed(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := IdentityWithYotiCallback(nil, nil, nil)(testAppData, w, r, &page.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.SelectYourIdentityOptions1, resp.Header.Get("Location"))
}
