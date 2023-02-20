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

type mockDataStore struct {
	data interface{}
	mock.Mock
}

func (m *mockDataStore) GetAll(ctx context.Context, pk string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, pk).Error(0)
}

func (m *mockDataStore) Get(ctx context.Context, pk, sk string, v interface{}) error {
	data, _ := json.Marshal(m.data)
	json.Unmarshal(data, v)
	return m.Called(ctx, pk, sk).Error(0)
}

func (m *mockDataStore) Put(ctx context.Context, pk, sk string, v interface{}) error {
	return m.Called(ctx, pk, sk, v).Error(0)
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

			dataStore := &mockDataStore{
				data: page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id", Identity: tc.identity},
			}
			dataStore.
				On("Get", r.Context(), "SHARECODE#a-share-code", "#METADATA#a-share-code").
				Return(nil)

			lpaStore := &mockLpaStore{}
			lpaStore.
				On("Get", mock.MatchedBy(func(ctx context.Context) bool {
					session := page.SessionDataFromContext(ctx)

					return assert.Equal(t, &page.SessionData{SessionID: "session-id", LpaID: "lpa-id"}, session)
				})).
				Return(&page.Lpa{}, nil)

			template := &mockTemplate{}
			template.
				On("Func", w, &startData{
					App:   testAppData,
					Start: page.Paths.CertificateProviderLogin + tc.query,
				}).
				Return(nil)

			err := Start(template.Func, lpaStore, dataStore)(testAppData, w, r)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			mock.AssertExpectationsForObjects(t, dataStore, lpaStore, template)
		})
	}
}

func TestStartWhenGettingShareCodeErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?share-code=a-share-code", nil)

	dataStore := &mockDataStore{
		data: page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id"},
	}
	dataStore.
		On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := Start(nil, nil, dataStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestStartWhenGettingLpaErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?share-code=a-share-code", nil)

	dataStore := &mockDataStore{
		data: page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id"},
	}
	dataStore.
		On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, expectedError)

	err := Start(nil, lpaStore, dataStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, dataStore, lpaStore)
}

func TestStartWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dataStore := &mockDataStore{
		data: page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id"},
	}
	dataStore.
		On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	lpaStore := &mockLpaStore{}
	lpaStore.
		On("Get", mock.Anything).
		Return(&page.Lpa{}, nil)

	template := &mockTemplate{}
	template.
		On("Func", mock.Anything, mock.Anything).
		Return(expectedError)

	err := Start(template.Func, lpaStore, dataStore)(testAppData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
