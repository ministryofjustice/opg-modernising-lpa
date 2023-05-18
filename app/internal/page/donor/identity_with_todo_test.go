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

func TestGetIdentityWithTodo(t *testing.T) {
	now := time.Now()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		}, nil)
	donorStore.
		On("Put", r.Context(), &page.Lpa{
			Donor: actor.Donor{FirstNames: "a", LastName: "b"},
			DonorIdentityUserData: identity.UserData{
				OK:          true,
				Provider:    identity.Passport,
				FirstNames:  "a",
				LastName:    "b",
				RetrievedAt: now,
			},
		}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithTodoData{
			App:            testAppData,
			IdentityOption: identity.Passport,
		}).
		Return(nil)

	err := IdentityWithTodo(template.Execute, donorStore, func() time.Time { return now }, identity.Passport)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithTodoWhenDonorStoreGetErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		}, expectedError)

	err := IdentityWithTodo(nil, donorStore, nil, identity.Passport)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithTodoWhenDonorStorePutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor: actor.Donor{FirstNames: "a", LastName: "b"},
		}, nil)
	donorStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := IdentityWithTodo(nil, donorStore, time.Now, identity.Passport)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostIdentityWithTodo(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			Donor: actor.Donor{FirstNames: "a", LastName: "b"},
			DonorIdentityUserData: identity.UserData{
				OK:          true,
				Provider:    identity.Passport,
				FirstNames:  "a",
				LastName:    "b",
				RetrievedAt: now,
			},
		}, nil)

	err := IdentityWithTodo(nil, donorStore, nil, identity.Passport)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.ReadYourLpa, resp.Header.Get("Location"))
}
