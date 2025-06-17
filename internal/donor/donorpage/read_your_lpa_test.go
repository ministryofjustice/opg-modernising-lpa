package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReadYourLpa(t *testing.T) {
	localizer := newMockLocalizer(t)

	testcases := map[string]struct {
		url        string
		bannerLang localize.Lang
	}{
		"en": {
			url:        "/?bannerLanguage=en",
			bannerLang: localize.En,
		},
		"cy": {
			url:        "/?bannerLanguage=cy",
			bannerLang: localize.Cy,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(tc.bannerLang).
				Return(localizer)

			bannerAppData := testAppData
			bannerAppData.Lang = tc.bannerLang
			bannerAppData.Localizer = localizer

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &readYourLpaData{
					App:       testAppData,
					BannerApp: bannerAppData,
					Donor:     &donordata.Provided{},
				}).
				Return(nil)

			err := ReadYourLpa(template.Execute, bundle)(testAppData, w, r, &donordata.Provided{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetReadYourLpaWhenNoBannerLanguage(t *testing.T) {
	cyAppData := testAppData
	cyAppData.Lang = localize.Cy

	testcases := map[string]struct {
		url            string
		appData        appcontext.Data
		bannerLanguage string
	}{
		"en missing": {
			url:            "/",
			appData:        testAppData,
			bannerLanguage: "en",
		},
		"cy missing": {
			url:            "/",
			appData:        cyAppData,
			bannerLanguage: "cy",
		},
		"en invalid": {
			url:            "/?bannerLanguage=blah",
			appData:        testAppData,
			bannerLanguage: "en",
		},
		"cy invalid": {
			url:            "/?bannerLanguage=blah",
			appData:        cyAppData,
			bannerLanguage: "cy",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, tc.url, nil)

			err := ReadYourLpa(nil, nil)(tc.appData, w, r, &donordata.Provided{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathReadYourLpa.Format("lpa-id")+"?bannerLanguage="+tc.bannerLanguage, resp.Header.Get("Location"))
		})
	}
}

func TestReadYourLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?bannerLanguage=en", nil)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ReadYourLpa(template.Execute, bundle)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}
