package certificateprovider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
)

func TestCertificateProviderPathString(t *testing.T) {
	assert.Equal(t, "/certificate-provider/{id}/anything", Path("/anything").String())
}

func TestCertificateProviderPathFormat(t *testing.T) {
	assert.Equal(t, "/certificate-provider/abc/anything", Path("/anything").Format("abc"))
}

func TestCertificateProviderPathRedirect(t *testing.T) {
	testcases := map[Path]string{
		Path("/something"):                "/certificate-provider/lpa-id/something",
		Path("/something?from=somewhere"): "/certificate-provider/lpa-id/somewhere",
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

func TestCertificateProviderCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		certificateProvider *certificateproviderdata.Provided
		url                 string
		expected            bool
	}{
		"empty path": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 "",
			expected:            false,
		},
		"unexpected path": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 "/whatever",
			expected:            true,
		},
		"unrestricted path": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 PathConfirmYourDetails.Format("123"),
			expected:            true,
		},
		"identity without task": {
			certificateProvider: &certificateproviderdata.Provided{},
			url:                 PathIdentityWithOneLogin.Format("123"),
			expected:            false,
		},
		"identity with task": {
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			url:      PathIdentityWithOneLogin.Format("123"),
			expected: true,
		},
		"provide certificate without task": {
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			url:      PathProvideCertificate.Format("123"),
			expected: false,
		},
		"provide certificate with task": {
			certificateProvider: &certificateproviderdata.Provided{
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.StateCompleted,
				},
			},
			url:      PathProvideCertificate.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CanGoTo(tc.certificateProvider, tc.url))
		})
	}
}
