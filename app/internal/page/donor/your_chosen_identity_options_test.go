package donor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &yourChosenIdentityOptionsData{
			App:            testAppData,
			IdentityOption: identity.Passport,
		}).
		Return(nil)

	err := YourChosenIdentityOptions(template.Execute)(testAppData, w, r, &page.Lpa{
		DonorIdentityOption: identity.Passport,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYourChosenIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := YourChosenIdentityOptions(template.Execute)(testAppData, w, r, &page.Lpa{
		DonorIdentityOption: identity.Passport,
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := YourChosenIdentityOptions(nil)(testAppData, w, r, &page.Lpa{
		DonorIdentityOption: identity.Passport,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.IdentityWithPassport, resp.Header.Get("Location"))
}
