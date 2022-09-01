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

func TestGetHowWouldYouLikeToBeContacted(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			Page: howWouldYouLikeToBeContactedPath,
			L:    localizer,
			Lang: En,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	HowWouldYouLikeToBeContacted(nil, localizer, En, template.Func, nil)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetHowWouldYouLikeToBeContactedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			Page: howWouldYouLikeToBeContactedPath,
			L:    localizer,
			Lang: En,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	HowWouldYouLikeToBeContacted(logger, localizer, En, template.Func, nil)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, logger)
}

func TestPostHowWouldYouLikeToBeContacted(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	dataStore := &mockDataStore{}
	dataStore.
		On("Save", []string{"email", "post"}).
		Return(nil)

	form := url.Values{
		"contact": {"email", "post"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	HowWouldYouLikeToBeContacted(nil, localizer, En, nil, dataStore)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, taskListPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestPostHowWouldYouLikeToBeContactedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			Page: howWouldYouLikeToBeContactedPath,
			L:    localizer,
			Lang: En,
			Errors: map[string]string{
				"contact": "selectContact",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	HowWouldYouLikeToBeContacted(nil, localizer, En, template.Func, nil)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadHowWouldYouLikeToBeContactedForm(t *testing.T) {
	form := url.Values{
		"contact": {"email", "phone"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readHowWouldYouLikeToBeContactedForm(r)

	assert.Equal(t, []string{"email", "phone"}, result.Contact)
}

func TestHowWouldYouLikeToBeContactedFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *howWouldYouLikeToBeContactedForm
		errors map[string]string
	}{
		"all": {
			form: &howWouldYouLikeToBeContactedForm{
				Contact: []string{"email", "phone", "text message", "post"},
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &howWouldYouLikeToBeContactedForm{},
			errors: map[string]string{
				"contact": "selectContact",
			},
		},
		"invalid": {
			form: &howWouldYouLikeToBeContactedForm{
				Contact: []string{"email", "what"},
			},
			errors: map[string]string{
				"contact": "selectContact",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
