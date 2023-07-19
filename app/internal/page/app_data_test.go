package page

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/localize"
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

			AppData{Lang: lang, LpaID: "lpa-id"}.Redirect(w, r, nil, "/dashboard")
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, url, resp.Header.Get("Location"))
		})
	}
}

func TestAppDataRedirectWhenCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		url      string
		lpa      *Lpa
		expected string
	}{
		"nil": {
			url:      "/",
			lpa:      nil,
			expected: Paths.HowToConfirmYourIdentityAndSign.Format("lpa-id"),
		},
		"nil and from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			lpa:      nil,
			expected: Paths.Restrictions.Format("lpa-id"),
		},
		"allowed": {
			url: "/",
			lpa: &Lpa{
				Type: LpaTypeHealthWelfare,
				Tasks: Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
					CheckYourLpa:               actor.TaskCompleted,
					PayForLpa:                  actor.PaymentTaskCompleted,
				},
			},
			expected: Paths.HowToConfirmYourIdentityAndSign.Format("lpa-id"),
		},
		"allowed from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			lpa:      &Lpa{Tasks: Tasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted}},
			expected: Paths.Restrictions.Format("lpa-id"),
		},
		"not allowed": {
			url:      "/",
			lpa:      &Lpa{},
			expected: Paths.TaskList.Format("lpa-id"),
		},
		"not allowed from": {
			url:      "/?from=" + Paths.Restrictions.Format("lpa-id"),
			lpa:      &Lpa{},
			expected: Paths.TaskList.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)
			w := httptest.NewRecorder()

			AppData{Lang: localize.En, LpaID: "lpa-id"}.Redirect(w, r, tc.lpa, Paths.HowToConfirmYourIdentityAndSign.Format("lpa-id"))
			resp := w.Result()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expected, resp.Header.Get("Location"))
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
