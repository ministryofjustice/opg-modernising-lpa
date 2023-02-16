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
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?share-code=a-share-code", nil)

	dataStore := &mockDataStore{
		data: page.ShareCodeData{LpaID: "lpa-id", SessionID: "session-id"},
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
			App:   appData,
			Start: page.Paths.CertificateProviderLogin + "?lpaId=lpa-id&sessionId=session-id",
		}).
		Return(nil)

	err := Start(template.Func, lpaStore, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, dataStore, lpaStore, template)
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

	err := Start(nil, nil, dataStore)(appData, w, r)

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

	err := Start(nil, lpaStore, dataStore)(appData, w, r)

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

	err := Start(template.Func, lpaStore, dataStore)(appData, w, r)

	assert.Equal(t, expectedError, err)
	mock.AssertExpectationsForObjects(t, lpaStore, template)
}
