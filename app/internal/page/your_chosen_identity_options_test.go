package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()

	selected := []IdentityOption{Passport, DwpAccount, UtilityBill}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			IdentityOptions: IdentityOptions{
				Selected: selected,
				First:    Passport,
				Second:   DwpAccount,
			},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourChosenIdentityOptionsData{
			App:          appData,
			Selected:     selected,
			FirstChoice:  Passport,
			SecondChoice: DwpAccount,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourChosenIdentityOptions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetYourChosenIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourChosenIdentityOptions(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetYourChosenIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			IdentityOptions: IdentityOptions{
				Selected: []IdentityOption{Passport, DwpAccount, UtilityBill},
			},
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourChosenIdentityOptions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			IdentityOptions: IdentityOptions{
				First:  Passport,
				Second: DwpAccount,
			},
		}, nil)

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := YourChosenIdentityOptions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityWithPassportPath, resp.Header.Get("Location"))
}
