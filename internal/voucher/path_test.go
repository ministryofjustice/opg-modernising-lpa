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
	r, _ := http.NewRequest(http.MethodGet, "/?from=/voucher/lpa-id/x", nil)
	w := httptest.NewRecorder()
	p := Path("/something")

	err := p.Redirect(w, r, appcontext.Data{Lang: localize.En}, "lpa-id")
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/voucher/lpa-id/x", resp.Header.Get("Location"))
}

func TestCanGoTo(t *testing.T) {
	testcases := map[string]struct {
		provided *voucherdata.Provided
		path     Path
		expected bool
	}{
		"unexpected path": {
			provided: &voucherdata.Provided{},
			path:     Path("/whatever"),
			expected: true,
		},
		"unrestricted path": {
			provided: &voucherdata.Provided{},
			path:     PathTaskList,
			expected: true,
		},
		"your name": {
			provided: &voucherdata.Provided{},
			path:     PathYourName,
			expected: true,
		},
		"your name when identity completed": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{ConfirmYourIdentity: task.StateCompleted},
			},
			path:     PathYourName,
			expected: false,
		},
		"verify donor details": {
			provided: &voucherdata.Provided{},
			path:     PathVerifyDonorDetails,
			expected: false,
		},
		"verify donor details when previous task completed": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted},
			},
			path:     PathVerifyDonorDetails,
			expected: true,
		},
		"verify donor details when already verified": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{ConfirmYourName: task.StateCompleted, VerifyDonorDetails: task.StateCompleted},
			},
			path:     PathVerifyDonorDetails,
			expected: false,
		},
		"confirm your identity": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName: task.StateCompleted,
				},
			},
			path:     PathConfirmYourIdentity,
			expected: false,
		},
		"confirm your identity when previous task completed": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:    task.StateCompleted,
					VerifyDonorDetails: task.StateCompleted,
				},
			},
			path:     PathConfirmYourIdentity,
			expected: true,
		},
		"sign the declaration": {
			provided: &voucherdata.Provided{
				Tasks: voucherdata.Tasks{
					ConfirmYourName:    task.StateCompleted,
					VerifyDonorDetails: task.StateCompleted,
				},
			},
			path:     PathSignTheDeclaration,
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
			path:     PathSignTheDeclaration,
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.path.CanGoTo(tc.provided))
		})
	}
}
