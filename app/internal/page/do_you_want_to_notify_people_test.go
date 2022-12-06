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

func TestGetDoYouWantToNotifyPeople(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App: appData,
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleFromStore(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{
			DoYouWantToNotifyPeople: "yes",
		}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App:          appData,
			WantToNotify: "yes",
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := DoYouWantToNotifyPeople(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetDoYouWantToNotifyPeopleWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App: appData,
		}).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostDoYouWantToNotifyPeople(t *testing.T) {
	testCases := []struct {
		WantToNotify     string
		ExistingAnswer   string
		ExpectedRedirect string
	}{
		{
			WantToNotify:     "yes",
			ExistingAnswer:   "no",
			ExpectedRedirect: appData.Paths.WhoShouldBeNotified,
		},
		{
			WantToNotify:     "no",
			ExistingAnswer:   "yes",
			ExpectedRedirect: appData.Paths.CheckYourLpa,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.WantToNotify, func(t *testing.T) {
			w := httptest.NewRecorder()

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.Anything, "session-id").
				Return(&Lpa{
					DoYouWantToNotifyPeople: tc.ExistingAnswer,
				}, nil)
			lpaStore.
				On("Put", mock.Anything, "session-id", &Lpa{
					DoYouWantToNotifyPeople: tc.WantToNotify,
				}).
				Return(nil)

			form := url.Values{
				"want-to-notify": {tc.WantToNotify},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", formUrlEncoded)

			err := DoYouWantToNotifyPeople(nil, lpaStore)(appData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.ExpectedRedirect, resp.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, lpaStore)
		})
	}
}

func TestPostDoYouWantToNotifyPeopleWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", &Lpa{DoYouWantToNotifyPeople: "yes"}).
		Return(expectedError)

	form := url.Values{
		"want-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := DoYouWantToNotifyPeople(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostDoYouWantToNotifyPeopleWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(&Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &doYouWantToNotifyPeopleData{
			App: appData,
			Errors: map[string]string{
				"want-to-notify": "selectDoYouWantToNotifyPeople",
			},
			Form: &doYouWantToNotifyPeopleForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("nope"))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := DoYouWantToNotifyPeople(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadDoYouWantToNotifyPeopleForm(t *testing.T) {
	form := url.Values{
		"want-to-notify": {"yes"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readDoYouWantToNotifyPeople(r)

	assert.Equal(t, "yes", result.WantToNotify)
}

func TestDoYouWantToNotifyPeopleFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *doYouWantToNotifyPeopleForm
		errors map[string]string
	}{
		"yes": {
			form: &doYouWantToNotifyPeopleForm{
				WantToNotify: "yes",
			},
			errors: map[string]string{},
		},
		"no": {
			form: &doYouWantToNotifyPeopleForm{
				WantToNotify: "no",
			},
			errors: map[string]string{},
		},
		"missing": {
			form: &doYouWantToNotifyPeopleForm{},
			errors: map[string]string{
				"want-to-notify": "selectDoYouWantToNotifyPeople",
			},
		},
		"invalid": {
			form: &doYouWantToNotifyPeopleForm{
				WantToNotify: "what",
			},
			errors: map[string]string{
				"want-to-notify": "selectDoYouWantToNotifyPeople",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
