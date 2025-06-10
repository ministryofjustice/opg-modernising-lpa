package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetReadTheLpa(t *testing.T) {
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
				Execute(w, &readTheLpaData{
					App:       testAppData,
					BannerApp: bannerAppData,
					Lpa:       &lpadata.Lpa{},
				}).
				Return(nil)

			err := ReadTheLpa(template.Execute, nil, bundle)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetReadTheLpaWhenNoBannerLanguage(t *testing.T) {
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

			err := ReadTheLpa(nil, nil, nil)(tc.appData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{
				LpaID: "lpa-id",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathReadTheLpa.Format("lpa-id")+"?bannerLanguage="+tc.bannerLanguage, resp.Header.Get("Location"))
		})
	}
}

func TestGetReadTheLpaWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?bannerLanguage=en", nil)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, nil, bundle)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?bannerLanguage=en", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), &attorneydata.Provided{
			LpaID: "lpa-id",
			Tasks: attorneydata.Tasks{
				ReadTheLpa: task.StateCompleted,
			},
		}).
		Return(nil)

	err := ReadTheLpa(nil, attorneyStore, nil)(testAppData, w, r, &attorneydata.Provided{LpaID: "lpa-id"}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, attorney.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}
