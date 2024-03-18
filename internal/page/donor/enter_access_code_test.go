package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterAccessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterAccessCodeData{
		App:  testAppData,
		Form: &enterAccessCodeForm{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	err := EnterAccessCode(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterAccessCodeOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterAccessCodeData{
		App:  testAppData,
		Form: &enterAccessCodeForm{},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(expectedError)

	err := EnterAccessCode(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCode(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCode := actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5", ActorUID: testUID}

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeDonor, "abcdef123456").
		Return(shareCode, nil)
	shareCodeStore.EXPECT().
		Linked(r.Context(), shareCode, "logged-in@example.com").
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(r.Context(), shareCode).
		Return(nil)

	err := EnterAccessCode(nil, shareCodeStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Dashboard.Format(), resp.Header.Get("Location"))
}

func TestPostEnterAccessCodeOnShareCodeStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {" abcdef123456  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeDonor, "abcdef123456").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, expectedError)

	err := EnterAccessCode(nil, shareCodeStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnShareCodeStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcde f-123456 "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterAccessCodeData{
		App:    testAppData,
		Form:   &enterAccessCodeForm{AccessCode: "abcdef123456", AccessCodeRaw: "abcde f-123456"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectAccessCode"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeDonor, "abcdef123456").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, dynamo.NotFoundError{})

	err := EnterAccessCode(template.Execute, shareCodeStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnDonorStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeDonor, "abcdef123456").
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, shareCodeStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnShareCodeStoreLinkedError(t *testing.T) {
	form := url.Values{
		"reference-number": {"abcdef123456"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	shareCodeStore := newMockShareCodeStore(t)
	shareCodeStore.EXPECT().
		Get(r.Context(), actor.TypeDonor, mock.Anything).
		Return(actor.ShareCodeData{LpaID: "lpa-id", SessionID: "aGV5"}, nil)
	shareCodeStore.EXPECT().
		Linked(r.Context(), mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(mock.Anything, mock.Anything).
		Return(nil)

	err := EnterAccessCode(nil, shareCodeStore, donorStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterAccessCodeOnValidationError(t *testing.T) {
	form := url.Values{
		"reference-number": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterAccessCodeData{
		App:    testAppData,
		Form:   &enterAccessCodeForm{},
		Errors: validation.With("reference-number", validation.EnterError{Label: "accessCode"}),
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, data).
		Return(nil)

	err := EnterAccessCode(template.Execute, nil, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateEnterAccessCodeForm(t *testing.T) {
	testCases := map[string]struct {
		form   *enterAccessCodeForm
		errors validation.List
	}{
		"valid": {
			form:   &enterAccessCodeForm{AccessCode: "abcdef123456"},
			errors: nil,
		},
		"too short": {
			form: &enterAccessCodeForm{AccessCode: "1"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "accessCode",
				Length: 12,
			}),
		},
		"too long": {
			form: &enterAccessCodeForm{AccessCode: "123456789ABCD"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "accessCode",
				Length: 12,
			}),
		},
		"empty": {
			form: &enterAccessCodeForm{},
			errors: validation.With("reference-number", validation.EnterError{
				Label: "accessCode",
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
