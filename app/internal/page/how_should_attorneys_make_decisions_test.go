package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{DecisionsDetails: "some decisions", DecisionsType: "jointly"}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App:              appData,
			DecisionsType:    "jointly",
			DecisionsDetails: "some decisions",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowShouldAttorneysMakeDecisionsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostHowShouldAttorneysMakeDecisions(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{DecisionsDetails: "", DecisionsType: ""}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{DecisionsDetails: "some decisions", DecisionsType: "jointly"}).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howShouldAttorneysMakeDecisionsData{
			App:              appData,
			DecisionsType:    "jointly",
			DecisionsDetails: "some decisions",
		}).
		Return(nil)

	form := url.Values{
		"decision-type": {"jointly"},
		"mixed-details": {"some decisions"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := HowShouldAttorneysMakeDecisions(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, wantReplacementAttorneysPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}
