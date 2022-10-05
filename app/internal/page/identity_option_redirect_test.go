package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityOptionRedirectFirst(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dataStore := &mockDataStore{
		data: Lpa{
			IdentityOptions: IdentityOptions{
				First:  Yoti,
				Second: DwpAccount,
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			IdentityOptions: IdentityOptions{
				First:   Yoti,
				Second:  DwpAccount,
				Current: 1,
			},
		}).
		Return(nil)

	err := IdentityOptionRedirect(dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityWithYotiPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetIdentityOptionRedirectSecond(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dataStore := &mockDataStore{
		data: Lpa{
			IdentityOptions: IdentityOptions{
				First:   Yoti,
				Second:  DwpAccount,
				Current: 1,
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)
	dataStore.
		On("Put", mock.Anything, "session-id", Lpa{
			IdentityOptions: IdentityOptions{
				First:   Yoti,
				Second:  DwpAccount,
				Current: 2,
			},
		}).
		Return(nil)

	err := IdentityOptionRedirect(dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, identityWithYotiPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}

func TestGetIdentityOptionRedirectFinal(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	dataStore := &mockDataStore{
		data: Lpa{
			IdentityOptions: IdentityOptions{
				First:   Yoti,
				Second:  DwpAccount,
				Current: 2,
			},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	err := IdentityOptionRedirect(dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, whatHappensWhenSigningPath, resp.Header.Get("Location"))
	mock.AssertExpectationsForObjects(t, dataStore)
}
