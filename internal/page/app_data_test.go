package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestAppDataRedirect(t *testing.T) {
	testCases := map[localize.Lang]string{
		localize.En: "/dashboard",
		localize.Cy: "/cy/dashboard",
	}

	for lang, url := range testCases {
		t.Run(lang.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			AppData{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, "/dashboard")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestAppDataContext(t *testing.T) {
	appData := AppData{LpaID: "me"}
	ctx := context.Background()

	assert.Equal(t, AppData{}, AppDataFromContext(ctx))
	assert.Equal(t, appData, AppDataFromContext(ContextWithAppData(ctx, appData)))
}

func TestIsDonor(t *testing.T) {
	appData := AppData{ActorType: actor.TypeDonor}
	assert.True(t, appData.IsDonor())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsDonor())
}

func TestIsCertificateProvider(t *testing.T) {
	appData := AppData{ActorType: actor.TypeCertificateProvider}
	assert.True(t, appData.IsCertificateProvider())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsCertificateProvider())
}

func TestIsAttorneyType(t *testing.T) {
	appData := AppData{ActorType: actor.TypeReplacementAttorney}
	assert.True(t, appData.IsAttorneyType())

	appData.ActorType = actor.TypeAttorney
	assert.True(t, appData.IsAttorneyType())

	appData.ActorType = actor.TypeTrustCorporation
	assert.True(t, appData.IsAttorneyType())

	appData.ActorType = actor.TypeReplacementTrustCorporation
	assert.True(t, appData.IsAttorneyType())

	appData.ActorType = actor.TypeCertificateProvider
	assert.False(t, appData.IsAttorneyType())
}

func TestIsReplacementAttorney(t *testing.T) {
	appData := AppData{ActorType: actor.TypeReplacementAttorney}
	assert.True(t, appData.IsReplacementAttorney())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsReplacementAttorney())
}

func TestIsTrustCorporation(t *testing.T) {
	assert.True(t, AppData{ActorType: actor.TypeTrustCorporation, AttorneyUID: actoruid.New()}.IsTrustCorporation())
	assert.True(t, AppData{ActorType: actor.TypeReplacementTrustCorporation, AttorneyUID: actoruid.New()}.IsTrustCorporation())

	assert.False(t, AppData{ActorType: actor.TypeAttorney, AttorneyUID: actoruid.New()}.IsTrustCorporation())
	assert.False(t, AppData{ActorType: actor.TypeReplacementAttorney, AttorneyUID: actoruid.New()}.IsTrustCorporation())
}

func TestAppDataIsAdmin(t *testing.T) {
	assert.True(t, AppData{SupporterData: &SupporterData{Permission: actor.PermissionAdmin}}.IsAdmin())
	assert.False(t, AppData{}.IsAdmin())
}

func TestAppDataEncodeQuery(t *testing.T) {
	testCases := map[string]struct {
		query               url.Values
		expectedQueryString string
	}{
		"with query": {
			query:               url.Values{"a": {"query"}, "b": {"string"}},
			expectedQueryString: "?a=query&b=string",
		},
		"without query": {
			expectedQueryString: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedQueryString, AppData{Query: tc.query}.EncodeQuery())
		})
	}
}
