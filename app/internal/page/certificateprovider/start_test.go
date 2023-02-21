package certificateprovider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestStart(t *testing.T) {
	testcases := map[string]struct {
		identity bool
		query    string
	}{
		"identity": {
			identity: true,
			query:    "?identity=1&lpaId=lpa-id&sessionId=session-id",
		},
		"sign in": {
			identity: false,
			query:    "?lpaId=lpa-id&sessionId=session-id",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?share-code=a-share-code", nil)

			dataStore := newMockDataStore(t)
			dataStore.
				ExpectGet(r.Context(), "SHARECODE#a-share-code", "#METADATA#a-share-code",
					page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: tc.identity}, nil)

			lpaStore := newMockLpaStore(t)
			lpaStore.
				On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					session := page.SessionDataFromContext(ctx)

					return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
				})).
				Return(&page.Lpa{}, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &startData{
					App:   testAppData,
					Start: page.Paths.CertificateProviderLogin + tc.query,
				}).
				Return(nil)

			err := Start(template.Execute, lpaStore, dataStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestStartWhenGettingShareCodeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?share-code=a-share-code", nil)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(mock.Anything, mock.Anything, mock.Anything,
			nil, expectedError)

	err := Start(nil, nil, dataStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestStartWhenGettingLpaErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?share-code=a-share-code", nil)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(mock.Anything, mock.Anything, mock.Anything,
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id"}, nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, expectedError)

	err := Start(nil, lpaStore, dataStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}

func TestStartWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dataStore := newMockDataStore(t)
	dataStore.
		ExpectGet(mock.Anything, mock.Anything, mock.Anything,
			page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id"}, nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", mock.Anything, mock.Anything).
		Return(expectedError)

	err := Start(template.Execute, lpaStore, dataStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
}
