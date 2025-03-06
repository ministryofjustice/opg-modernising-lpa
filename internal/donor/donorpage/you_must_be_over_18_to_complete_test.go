package donorpage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/stretchr/testify/assert"
)

func TestGetYouMustBeOver18ToComplete(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{
			DateOfBirth: date.Today().AddDate(-18, 0, 1),
		},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, guidanceData{App: testAppData, Donor: donor}).
		Return(expectedError)

	err := YouMustBeOver18ToComplete(template.Execute)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetMustBeOver18ToCompleteWithoutUnder18Donor(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YouMustBeOver18ToComplete(nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{
			DateOfBirth: date.Today().AddDate(-18, 0, 0),
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}
