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

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"

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

func TestGetEnterReferenceCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceCodeData{
		App:  testAppData,
		Form: &enterReferenceCodeForm{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := EnterReferenceCode(template.Execute, newMockLpaStore(t), newMockDataStore(t))(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterReferenceCodeOnTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	data := enterReferenceCodeData{
		App:  testAppData,
		Form: &enterReferenceCodeForm{},
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(expectedError)

	err := EnterReferenceCode(template.Execute, newMockLpaStore(t), newMockDataStore(t))(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceCode(t *testing.T) {
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
				"reference-code": {"a-ref-code"},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			dataStore := newMockDataStore(t)
			dataStore.
				ExpectGet(r.Context(), "SHARECODE#a-ref-code", "#METADATA#a-ref-code",
					page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: tc.Identity}, nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					session := page.SessionDataFromContext(ctx)

					return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
				})).
				Return(&page.Lpa{}, nil)

			err := EnterReferenceCode(nil, lpaStore, dataStore)(testAppData, w, r)

			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.CertificateProviderLogin+"?"+tc.ExpectedQuery, resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterReferenceCodeOnDataStoreError(t *testing.T) {
	form := url.Values{
		"reference-code": {"  a-ref-code  "},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(r.Context(), "SHARECODE#a-ref-code", "#METADATA#a-ref-code",
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: true}, expectedError)

	err := EnterReferenceCode(nil, nil, dataStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceCodeOnDataStoreNotFoundError(t *testing.T) {
	form := url.Values{
		"reference-code": {"a-ref-code"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceCodeData{
		App:    testAppData,
		Form:   &enterReferenceCodeForm{ReferenceCode: "a-ref-code"},
		Errors: validation.With("reference-code", validation.CustomError{Label: "incorrectReferenceCode"}),
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(r.Context(), "SHARECODE#a-ref-code", "#METADATA#a-ref-code",
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: true}, dynamo.NotFoundError{})

	err := EnterReferenceCode(template.Execute, nil, dataStore)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceCodeOnLpaStoreError(t *testing.T) {
	form := url.Values{
		"reference-code": {"a-ref-code"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(r.Context(), "SHARECODE#a-ref-code", "#METADATA#a-ref-code",
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: true}, nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", mock.MatchedBy(func(ctx context.Context) bool {
			session := page.SessionDataFromContext(ctx)

			return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
		})).
		Return(&page.Lpa{}, expectedError)

	err := EnterReferenceCode(nil, lpaStore, dataStore)(testAppData, w, r)

	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterReferenceCodeOnValidationError(t *testing.T) {
	form := url.Values{
		"reference-code": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	data := enterReferenceCodeData{
		App:    testAppData,
		Form:   &enterReferenceCodeForm{},
		Errors: validation.With("reference-code", validation.EnterError{Label: "referenceCode"}),
	}

	template := newMockTemplate(t)
	template.
		On("Execute", w, data).
		Return(nil)

	err := EnterReferenceCode(template.Execute, nil, nil)(testAppData, w, r)

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
