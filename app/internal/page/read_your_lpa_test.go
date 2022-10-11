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

func TestGetReadYourLpa(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &readYourLpaData{
			App:  appData,
			Form: &readYourLpaForm{},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ReadYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetReadYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ReadYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetReadYourLpaFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	lpa := Lpa{
		CheckedAgain:    true,
		ConfirmFreeWill: true,
		SignatureCode:   "1234",
	}

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(lpa, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &readYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &readYourLpaForm{
				Checked:   true,
				Confirm:   true,
				Signature: "1234",
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ReadYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestPostReadYourLpa(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", Lpa{
			Tasks: Tasks{
				ConfirmYourIdentityAndSign: TaskCompleted,
			},
			CheckedAgain:    true,
			ConfirmFreeWill: true,
			SignatureCode:   "1234",
		}).
		Return(nil)

	form := url.Values{
		"checked":   {"1"},
		"confirm":   {"1"},
		"signature": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ReadYourLpa(nil, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, signingConfirmationPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostReadYourLpaWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)
	lpaStore.
		On("Put", mock.Anything, "session-id", mock.Anything).
		Return(expectedError)

	form := url.Values{
		"checked":   {"1"},
		"confirm":   {"1"},
		"signature": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ReadYourLpa(nil, lpaStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostReadYourLpaWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything, "session-id").
		Return(Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.MatchedBy(func(data *readYourLpaData) bool {
			return assert.Equal(t, map[string]string{"confirm": "selectConfirmMadeThisLpaOfOwnFreeWill"}, data.Errors)
		})).
		Return(nil)

	form := url.Values{
		"checked":   {"1"},
		"signature": {"1234"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	err := ReadYourLpa(template.Func, lpaStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestReadReadYourLpaForm(t *testing.T) {
	assert := assert.New(t)

	form := url.Values{
		"checked": {" 1   "},
		"happy":   {" 0"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", formUrlEncoded)

	result := readReadYourLpaForm(r)

	assert.Equal(true, result.Checked)
	assert.Equal(false, result.Confirm)
}

func TestReadYourLpaFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *readYourLpaForm
		errors map[string]string
	}{
		"valid": {
			form: &readYourLpaForm{
				Confirm:   true,
				Checked:   true,
				Signature: "1234",
			},
			errors: map[string]string{},
		},
		"invalid-all": {
			form: &readYourLpaForm{},
			errors: map[string]string{
				"checked":   "selectReadAndCheckedLpa",
				"confirm":   "selectConfirmMadeThisLpaOfOwnFreeWill",
				"signature": "enterSignatureCode",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
