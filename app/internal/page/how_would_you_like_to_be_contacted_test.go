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

func TestGetHowWouldYouLikeToBeContacted(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			App: appData,
		}).
		Return(nil)

	err := HowWouldYouLikeToBeContacted(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetHowWouldYouLikeToBeContactedWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, expectedError)

	err := HowWouldYouLikeToBeContacted(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetHowWouldYouLikeToBeContactedFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{Contact: []string{"email"}}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			App:     appData,
			Contact: []string{"email"},
		}).
		Return(nil)

	err := HowWouldYouLikeToBeContacted(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetHowWouldYouLikeToBeContactedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			App: appData,
		}).
		Return(expectedError)

	err := HowWouldYouLikeToBeContacted(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostHowWouldYouLikeToBeContacted(t *testing.T) {
	form := url.Values{
		"contact": {"email", "post"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{Contact: []string{"email", "post"}}).
		Return(nil)

	err := HowWouldYouLikeToBeContacted(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+Paths.TaskList, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowWouldYouLikeToBeContactedWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"contact": {"email", "post"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &Lpa{Contact: []string{"email", "post"}}).
		Return(expectedError)

	err := HowWouldYouLikeToBeContacted(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostHowWouldYouLikeToBeContactedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", formUrlEncoded)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &howWouldYouLikeToBeContactedData{
			App: appData,
			Errors: map[string]string{
				"contact": "selectContact",
			},
		}).
		Return(nil)

	err := HowWouldYouLikeToBeContacted(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
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
