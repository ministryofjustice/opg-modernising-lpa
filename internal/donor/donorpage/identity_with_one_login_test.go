package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdentityWithOneLogin(t *testing.T) {
	for country, confidenceLevel := range map[string]onelogin.ConfidenceLevel{
		"GB":    onelogin.ConfidenceLevelLow,
		"Other": onelogin.ConfidenceLevelMedium,
	} {
		t.Run(country, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			client := newMockOneLoginClient(t)
			client.EXPECT().
				AuthCodeURL("i am random", "i am random", "cy", confidenceLevel).
				Return("http://auth", nil)

			sessionStore := newMockSessionStore(t)
			sessionStore.EXPECT().
				SetOneLogin(r, w, &sesh.OneLoginSession{State: "i am random", Nonce: "i am random", Locale: "cy", Redirect: donor.PathIdentityWithOneLoginCallback.Format("lpa-id"), LpaID: "lpa-id"}).
				Return(nil)

			err := IdentityWithOneLogin(client, sessionStore, func(int) string { return "i am random" })(appcontext.Data{Lang: localize.Cy}, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{
					Address: place.Address{
						Country: country,
					},
				},
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "http://auth", resp.Header.Get("Location"))
		})
	}
}

func TestIdentityWithOneLoginWhenAuthCodeURLError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("http://auth?locale=en", expectedError)

	err := IdentityWithOneLogin(client, nil, func(int) string { return "i am random" })(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIdentityWithOneLoginWhenStoreSaveError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	client := newMockOneLoginClient(t)
	client.EXPECT().
		AuthCodeURL(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("http://auth?locale=en", nil)

	sessionStore := newMockSessionStore(t)
	sessionStore.EXPECT().
		SetOneLogin(r, w, mock.Anything).
		Return(expectedError)

	err := IdentityWithOneLogin(client, sessionStore, func(int) string { return "i am random" })(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
