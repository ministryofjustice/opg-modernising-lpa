package attorney

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
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
		Path("/something"):                "/attorney/lpa-id/something",
		Path("/something?from=somewhere"): "/attorney/lpa-id/somewhere",
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
		Path("/something"):                "/attorney/lpa-id/something",
		Path("/something?from=somewhere"): "/attorney/lpa-id/somewhere",
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
		url      string
		expected bool
	}{
		"empty path": {
			attorney: &attorneydata.Provided{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			attorney: &attorneydata.Provided{},
			url:      "/whatever",
			expected: true,
		},
		"unrestricted path": {
			attorney: &attorneydata.Provided{},
			url:      PathConfirmYourDetails.Format("123"),
			expected: true,
		},
		"sign without task": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			url:      PathSign.Format("123"),
			expected: false,
		},
		"sign with task": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			url:      PathSign.Format("123"),
			expected: true,
		},
		"would like second signatory not trust corp": {
			attorney: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			url:      PathWouldLikeSecondSignatory.Format("123"),
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
			url:      PathWouldLikeSecondSignatory.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CanGoTo(tc.attorney, tc.url))
		})
	}
}
