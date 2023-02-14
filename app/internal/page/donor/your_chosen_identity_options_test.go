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

func TestGetYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			IdentityOption: identity.Passport,
		}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &yourChosenIdentityOptionsData{
			App:            page.TestAppData,
			IdentityOption: identity.Passport,
		}).
		Return(nil)

	err := YourChosenIdentityOptions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetYourChosenIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := YourChosenIdentityOptions(nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetYourChosenIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			IdentityOption: identity.Passport,
		}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(page.ExpectedError)

	err := YourChosenIdentityOptions(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{
			IdentityOption: identity.Passport,
		}, nil)

	err := YourChosenIdentityOptions(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithPassport, resp.Header.Get("Location"))
}
