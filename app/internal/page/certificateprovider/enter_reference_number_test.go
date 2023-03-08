package certificateprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *mockDataStore) ExpectGet(ctx, pk, sk, data interface{}, err error) {
	m.
		On("Get", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func TestGetEnterReferenceNumber(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceNumberData{
		App:  testAppData,
		Form: &enterReferenceNumberForm{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, newMockLpaStore(t), newMockDataStore(t))(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReferenceNumberOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceNumberData{
		App:  testAppData,
		Form: &enterReferenceNumberForm{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(expectedError)

	err := EnterReferenceNumber(template.Execute, newMockLpaStore(t), newMockDataStore(t))(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumber(t *testing.T) {
	testCases := map[string]struct {
		Identity      bool
		ExpectedQuery string
	}{
		"with identity": {
			Identity:      true,
			ExpectedQuery: "identity=1&lpaId=lpa-id&sessionId=session-id",
		},
		"without identity": {
			Identity:      false,
			ExpectedQuery: "lpaId=lpa-id&sessionId=session-id",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			form := url.Values{
				"reference-number": {"aRefNumber12"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			dataStore := newMockDataStore(t)
			dataStore.
				ExpectGet(r.Context(), "SHARECODE#aRefNumber12", "#METADATA#aRefNumber12",
					page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: tc.Identity}, nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					session := page.SessionDataFromContext(ctx)

					return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
				})).
				Return(&page.Lpa{}, nil)

			err := EnterReferenceNumber(nil, lpaStore, dataStore)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProviderLogin+"?"+tc.ExpectedQuery, resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceNumberOnDataStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"  aRefNumber12  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(r.Context(), "SHARECODE#aRefNumber12", "#METADATA#aRefNumber12",
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: true}, expectedError)

	err := EnterReferenceNumber(nil, nil, dataStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnDataStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-number": {"aRefNumber12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{ReferenceNumber: "aRefNumber12"},
		Errors: validation.With("reference-number", validation.CustomError{Label: "incorrectReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(r.Context(), "SHARECODE#aRefNumber12", "#METADATA#aRefNumber12",
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: true}, dynamo.NotFoundError{})

	err := EnterReferenceNumber(template.Execute, nil, dataStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnLpaStoreError(t *testing.T) {
	form := url.Values{
		"reference-number": {"aRefNumber12"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(r.Context(), "SHARECODE#aRefNumber12", "#METADATA#aRefNumber12",
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: true}, nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := page.SessionDataFromContext(ctx)

			return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
		})).
		Return(&page.Lpa{}, expectedError)

	err := EnterReferenceNumber(nil, lpaStore, dataStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceNumberOnValidationError(t *testing.T) {
	form := url.Values{
		"reference-number": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceNumberData{
		App:    testAppData,
		Form:   &enterReferenceNumberForm{},
		Errors: validation.With("reference-number", validation.EnterError{Label: "twelveCharactersReferenceNumber"}),
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := EnterReferenceNumber(template.Execute, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateEnterReferenceNumberForm(t *testing.T) {
	testCases := map[string]struct {
		form   *enterReferenceNumberForm
		errors validation.List
	}{
		"valid": {
			form:   &enterReferenceNumberForm{ReferenceNumber: "abcdef123456"},
			errors: nil,
		},
		"too short": {
			form: &enterReferenceNumberForm{ReferenceNumber: "1"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "referenceNumberMustBeTwelveCharacters",
				Length: 12,
			}),
		},
		"too long": {
			form: &enterReferenceNumberForm{ReferenceNumber: "abcdef1234567"},
			errors: validation.With("reference-number", validation.StringLengthError{
				Label:  "referenceNumberMustBeTwelveCharacters",
				Length: 12,
			}),
		},
		"empty": {
			form: &enterReferenceNumberForm{},
			errors: validation.With("reference-number", validation.EnterError{
				Label: "twelveCharactersReferenceNumber",
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}

}
