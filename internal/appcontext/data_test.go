package appcontext

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

			Data{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, "/dashboard")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestAppDataContext(t *testing.T) {
	appData := Data{LpaID: "me"}
	ctx := context.Background()

	assert.Equal(t, Data{}, DataFromContext(ctx))
	assert.Equal(t, appData, DataFromContext(ContextWithData(ctx, appData)))
}

func TestIsDonor(t *testing.T) {
	appData := Data{ActorType: actor.TypeDonor}
	assert.True(t, appData.IsDonor())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsDonor())
}

func TestIsCertificateProvider(t *testing.T) {
	appData := Data{ActorType: actor.TypeCertificateProvider}
	assert.True(t, appData.IsCertificateProvider())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsCertificateProvider())
}

func TestIsAttorneyType(t *testing.T) {
	appData := Data{ActorType: actor.TypeReplacementAttorney}
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
	appData := Data{ActorType: actor.TypeReplacementAttorney}
	assert.True(t, appData.IsReplacementAttorney())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsReplacementAttorney())
}

func TestIsTrustCorporation(t *testing.T) {
	assert.True(t, Data{ActorType: actor.TypeTrustCorporation, AttorneyUID: actoruid.New()}.IsTrustCorporation())
	assert.True(t, Data{ActorType: actor.TypeReplacementTrustCorporation, AttorneyUID: actoruid.New()}.IsTrustCorporation())

	assert.False(t, Data{ActorType: actor.TypeAttorney, AttorneyUID: actoruid.New()}.IsTrustCorporation())
	assert.False(t, Data{ActorType: actor.TypeReplacementAttorney, AttorneyUID: actoruid.New()}.IsTrustCorporation())
}

func TestAppDataIsAdmin(t *testing.T) {
	assert.True(t, Data{SupporterData: &SupporterData{Permission: actor.PermissionAdmin}}.IsAdmin())
	assert.False(t, Data{}.IsAdmin())
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
			assert.Equal(t, tc.expectedQueryString, Data{Query: tc.query}.EncodeQuery())
		})
	}
}
