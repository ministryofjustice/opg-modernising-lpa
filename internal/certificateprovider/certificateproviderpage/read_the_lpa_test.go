package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
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

			donor := &lpadata.Lpa{}

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(tc.bannerLang).
				Return(localizer)

			bannerAppData := testAppData
			bannerAppData.Lang = tc.bannerLang
			bannerAppData.Localizer = localizer

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &readTheLpaData{App: testAppData, BannerApp: bannerAppData, Lpa: donor}).
				Return(nil)

			err := ReadTheLpa(template.Execute, nil, bundle)(testAppData, w, r, nil, donor)
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

			err := ReadTheLpa(nil, nil, nil)(tc.appData, w, r, nil, &lpadata.Lpa{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, certificateprovider.PathReadTheLpa.FormatQuery("lpa-id", url.Values{
				"bannerLanguage": {tc.bannerLanguage},
			}), resp.Header.Get("Location"))
		})
	}
}

func TestGetReadTheLpaWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?bannerLanguage=en", nil)

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.En).
		Return(testAppData.Localizer)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, nil, bundle)(testAppData, w, r, nil, nil)

	assert.Equal(t, expectedError, err)
}

func TestPostReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?bannerLanguage=en", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), &certificateproviderdata.Provided{
			Tasks: certificateproviderdata.Tasks{
				ReadTheLpa: task.StateCompleted,
			},
		}).
		Return(nil)

	err := ReadTheLpa(nil, certificateProviderStore, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{
		LpaID:                            "lpa-id",
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Paid:                             true,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, certificateprovider.PathWhatHappensNext.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostReadTheLpaWithAttorneyWhenCertificateStorePutErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?bannerLanguage=en", nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := ReadTheLpa(nil, certificateProviderStore, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{
		LpaID:                            "lpa-id",
		SignedAt:                         time.Now(),
		WitnessedByCertificateProviderAt: time.Now(),
		Paid:                             true,
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
