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

	dataStore := &mockDataStore{data: Lpa{IdentityOptions: selected}}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &yourChosenIdentityOptionsData{
			App:           appData,
			ArticleLabels: identityOptionArticleLabels,
			Labels:        identityOptionLabels,
			Selected:      selected,
			FirstChoice:   Passport,
			SecondChoice:  DwpAccount,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourChosenIdentityOptions(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetYourChosenIdentityOptionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourChosenIdentityOptions(nil, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetYourChosenIdentityOptionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{data: Lpa{IdentityOptions: []IdentityOption{Passport, DwpAccount, UtilityBill}}}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YourChosenIdentityOptions(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostYourChosenIdentityOptions(t *testing.T) {
	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	err := YourChosenIdentityOptions(nil, nil)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityWithEasyIDPath, resp.Header.Get("Location"))
}
