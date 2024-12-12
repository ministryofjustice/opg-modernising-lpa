package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyPathString(t *testing.T) {
	assert.Equal(t, "/attorney/{id}/anything", Path("/anything").String())
}

func TestAttorneyPathFormat(t *testing.T) {
	assert.Equal(t, "/attorney/abc/anything", Path("/anything").Format("abc"))
}

func TestAttorneyPathRedirect(t *testing.T) {
	testcases := map[Path]string{
		Path("/something"): "/attorney/lpa-id/something",
		Path("/something?from=/attorney/lpa-id/somewhere"): "/attorney/lpa-id/somewhere",
	}

	for path, expectedURL := range testcases {
		t.Run(path.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, path.Format("lpa-id"), nil)
			w := httptest.NewRecorder()

			err := path.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, expectedURL, resp.Header.Get("Location"))
		})
	}
}

func TestAttorneyPathRedirectQuery(t *testing.T) {
	testcases := map[Path]string{
		Path("/something"): "/attorney/lpa-id/something",
		Path("/something?from=/attorney/lpa-id/somewhere"): "/attorney/lpa-id/somewhere",
	}

	for path, expectedURL := range testcases {
		t.Run(path.String(), func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, path.Format("lpa-id"), nil)
			w := httptest.NewRecorder()

			err := path.RedirectQuery(w, r, appcontext.Data{Lang: localize.En}, "lpa-id", url.Values{"q": {"1"}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, expectedURL+"?q=1", resp.Header.Get("Location"))
		})
	}
}

func TestAttorneyCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		attorney *attorneydata.Provided
		path     Path
		expected bool
	}{
		"empty path": {
			attorney: &attorneydata.Provided{},
			path:     Path(""),
			expected: true,
		},
		"unrestricted path": {
			attorney: &attorneydata.Provided{},
			path:     PathConfirmYourDetails,
			expected: true,
		},
		"sign without task": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			path:     PathSign,
			expected: false,
		},
		"sign with task": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			path:     PathSign,
			expected: true,
		},
		"would like second signatory not trust corp": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			path:     PathWouldLikeSecondSignatory,
			expected: false,
		},
		"would like second signatory as trust corp": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
				IsTrustCorporation: true,
			},
			path:     PathWouldLikeSecondSignatory,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			signedAt := time.Now()

			assert.Equal(t, tc.expected, tc.path.CanGoTo(tc.attorney, &lpadata.Lpa{
				Paid:                             true,
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				CertificateProvider: lpadata.CertificateProvider{
					SignedAt: &signedAt,
				},
			}))
		})
	}
}
