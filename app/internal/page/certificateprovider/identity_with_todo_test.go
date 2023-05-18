package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetIdentityWithTodo(t *testing.T) {
	now := time.Now()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{FirstNames: "a", LastName: "b"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), &actor.CertificateProviderProvidedDetails{
			FirstNames: "a",
			LastName:   "b",
			IdentityUserData: identity.UserData{
				OK:          true,
				Provider:    identity.Passport,
				FirstNames:  "a",
				LastName:    "b",
				RetrievedAt: now,
			},
		}).
		Return(nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &identityWithTodoData{
			App:            testAppData,
			IdentityOption: identity.Passport,
		}).
		Return(nil)

	err := IdentityWithTodo(template.Execute, func() time.Time { return now }, identity.Passport, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetIdentityWithTodoWhenCertificateProviderStoreGetErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, expectedError)

	err := IdentityWithTodo(nil, nil, identity.Passport, certificateProviderStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestGetIdentityWithTodoWhenDonorStorePutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{FirstNames: "a", LastName: "b"}, nil)
	certificateProviderStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := IdentityWithTodo(nil, time.Now, identity.Passport, certificateProviderStore)(testAppData, w, r)
	assert.Equal(t, expectedError, err)
}

func TestPostIdentityWithTodo(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.
		On("Get", r.Context()).
		Return(&actor.CertificateProviderProvidedDetails{}, nil)

	err := IdentityWithTodo(nil, nil, identity.Passport, certificateProviderStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderReadTheLpa, resp.Header.Get("Location"))
}
