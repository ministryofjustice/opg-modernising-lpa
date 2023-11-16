package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

func TestAppDataBuildUrl(t *testing.T) {
	testCases := map[string]struct {
		lang localize.Lang
		url  string
		want string
	}{
		"english":        {lang: localize.En, url: "/example.org", want: "/example.org"},
		"welsh":          {lang: localize.Cy, url: "/example.org", want: "/cy/example.org"},
		"other language": {lang: localize.Lang(3), url: "/example.org", want: "/example.org"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			builtUrl := AppData{Lang: tc.lang}.BuildUrl(tc.url)
			assert.Equal(t, tc.want, builtUrl)
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

func TestIsReplacementAttorney(t *testing.T) {
	appData := AppData{ActorType: actor.TypeReplacementAttorney}
	assert.True(t, appData.IsReplacementAttorney())

	appData.ActorType = actor.TypeAttorney
	assert.False(t, appData.IsReplacementAttorney())
}

func TestIsTrustCorporation(t *testing.T) {
	assert.True(t, AppData{ActorType: actor.TypeAttorney}.IsTrustCorporation())
	assert.True(t, AppData{ActorType: actor.TypeReplacementAttorney}.IsTrustCorporation())

	assert.False(t, AppData{ActorType: actor.TypeAttorney, AttorneyID: "1"}.IsTrustCorporation())
	assert.False(t, AppData{ActorType: actor.TypeReplacementAttorney, AttorneyID: "1"}.IsTrustCorporation())
}
