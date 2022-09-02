package page

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaType(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			Page: lpaTypePath,
			L:    localizer,
			Lang: En,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	LpaType(nil, localizer, En, template.Func, dataStore)(w, withSession(r))
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetLpaTypeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			Page: lpaTypePath,
			L:    localizer,
			Lang: En,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	LpaType(logger, localizer, En, template.Func, dataStore)(w, withSession(r))
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, logger)
}

func TestPostLpaType(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{Type: "pfa"}).
		Return(nil)

	form := url.Values{
		"lpa-type": {"pfa"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	LpaType(nil, localizer, En, nil, dataStore)(w, withSession(r))
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whoIsTheLpaForPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostLpaTypeWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id", mock.Anything).
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &lpaTypeData{
			Page: lpaTypePath,
			L:    localizer,
			Lang: En,
			Errors: map[string]string{
				"lpa-type": "selectLpaType",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	LpaType(nil, localizer, En, template.Func, dataStore)(w, withSession(r))
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadLpaTypeForm(t *testing.T) {
	form := url.Values{
		"lpa-type": {"pfa"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readLpaTypeForm(r)

	assert.Equal(t, "pfa", result.LpaType)
}

func TestLpaTypeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lpaTypeForm
		errors map[string]string
	}{
		"pfa": {
			form: &lpaTypeForm{
				LpaType: "pfa",
			},
			errors: map[string]string{},
		},
		"hw": {
			form: &lpaTypeForm{
				LpaType: "hw",
			},
			errors: map[string]string{},
		},
		"both": {
			form: &lpaTypeForm{
				LpaType: "both",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &lpaTypeForm{},
			errors: map[string]string{
				"lpa-type": "selectLpaType",
			},
		},
		"invalid": {
			form: &lpaTypeForm{
				LpaType: "what",
			},
			errors: map[string]string{
				"lpa-type": "selectLpaType",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
