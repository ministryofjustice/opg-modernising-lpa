package voucher

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
)

func TestPathString(t *testing.T) {
	assert.Equal(t, "/voucher/{id}/anything", Path("/anything").String())
}

func TestPathFormat(t *testing.T) {
	assert.Equal(t, "/voucher/abc/anything", Path("/anything").Format("abc"))
}

func TestPathRedirect(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, p.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPathRedirectWhenFrom(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/?from=/x", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/x", resp.Header.Get("Location"))
}

func TestCanGoTo(t *testing.T) {
	testcases := map[string]struct {
		provided *voucherdata.Provided
		url      string
		expected bool
	}{
		"empty path": {
			provided: &voucherdata.Provided{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			provided: &voucherdata.Provided{},
			url:      "/whatever",
			expected: true,
		},
		"unrestricted path": {
			provided: &voucherdata.Provided{},
			url:      PathTaskList.Format("123"),
			expected: true,
		},
		"verify donor details": {
			provided: &voucherdata.Provided{},
			url:      PathVerifyDonorDetails.Format("123"),
			expected: false,
		},
		"verify donor details when previous task completed": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted},
			},
			url:      PathVerifyDonorDetails.Format("123"),
			expected: true,
		},
		"verify donor details when already verified": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted, VerifyDonorDetails: task.StateCompleted},
			},
			url:      PathVerifyDonorDetails.Format("123"),
			expected: false,
		},
		"confirm your identity": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName: task.StateCompleted,
				},
			},
			url:      PathConfirmYourIdentity.Format("123"),
			expected: false,
		},
		"confirm your identity when previous task completed": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:    task.StateCompleted,
					VerifyDonorDetails: task.StateCompleted,
				},
			},
			url:      PathConfirmYourIdentity.Format("123"),
			expected: true,
		},
		"sign the declaration": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:    task.StateCompleted,
					VerifyDonorDetails: task.StateCompleted,
				},
			},
			url:      PathSignTheDeclaration.Format("123"),
			expected: false,
		},
		"sign the declaration when previous task completed": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:     task.StateCompleted,
					VerifyDonorDetails:  task.StateCompleted,
					ConfirmYourIdentity: task.StateCompleted,
				},
			},
			url:      PathSignTheDeclaration.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, CanGoTo(tc.provided, tc.url))
		})
	}
}
