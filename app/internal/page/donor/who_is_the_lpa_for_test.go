package donor

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWhoIsTheLpaFor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: page.TestAppData,
		}).
		Return(nil)

	err := WhoIsTheLpaFor(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, lpaStore)
}

func TestGetWhoIsTheLpaForWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, page.ExpectedError)

	err := WhoIsTheLpaFor(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestGetWhoIsTheLpaForFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{WhoFor: "me"}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App:    page.TestAppData,
			WhoFor: "me",
		}).
		Return(nil)

	err := WhoIsTheLpaFor(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetWhoIsTheLpaForWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App: page.TestAppData,
		}).
		Return(page.ExpectedError)

	err := WhoIsTheLpaFor(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Equal(t, page.ExpectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestPostWhoIsTheLpaFor(t *testing.T) {
	f := url.Values{
		"who-for": {"me"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{WhoFor: "me"}).
		Return(nil)

	err := WhoIsTheLpaFor(nil, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/lpa/lpa-id"+page.Paths.LpaType, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhoIsTheLpaForWhenStoreErrors(t *testing.T) {
	f := url.Values{
		"who-for": {"me"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)
	lpaStore.
		On("Put", r.Context(), &page.Lpa{WhoFor: "me"}).
		Return(page.ExpectedError)

	err := WhoIsTheLpaFor(nil, lpaStore)(page.TestAppData, w, r)

	assert.Equal(t, page.ExpectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore)
}

func TestPostWhoIsTheLpaForWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	lpaStore := &page.MockLpaStore{}
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{}, nil)

	template := &page.MockTemplate{}
	template.
		On("Func", w, &whoIsTheLpaForData{
			App:    page.TestAppData,
			Errors: validation.With("who-for", validation.SelectError{Label: "whoTheLpaIsFor"}),
		}).
		Return(nil)

	err := WhoIsTheLpaFor(template.Func, lpaStore)(page.TestAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestReadWhoIsTheLpaForForm(t *testing.T) {
	f := url.Values{
		"who-for": {"me"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readWhoIsTheLpaForForm(r)

	assert.Equal(t, "me", result.WhoFor)
}

func TestWhoIsTheLpaForFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *whoIsTheLpaForForm
		errors validation.List
	}{
		"me": {
			form: &whoIsTheLpaForForm{
				WhoFor: "me",
			},
		},
		"someone-else": {
			form: &whoIsTheLpaForForm{
				WhoFor: "someone-else",
			},
		},
		"missing": {
			form:   &whoIsTheLpaForForm{},
			errors: validation.With("who-for", validation.SelectError{Label: "whoTheLpaIsFor"}),
		},
		"invalid": {
			form: &whoIsTheLpaForForm{
				WhoFor: "what",
			},
			errors: validation.With("who-for", validation.SelectError{Label: "whoTheLpaIsFor"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
