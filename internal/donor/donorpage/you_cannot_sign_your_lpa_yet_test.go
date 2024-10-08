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

func TestGetYouCannotSignYourLpaYetWithUnder18Actors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{DateOfBirth: date.New(date.Today().YearString(), "1", "2")},
		}},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, guidanceData{App: testAppData, Donor: donor}).
		Return(nil)

	err := YouCannotSignYourLpaYet(template.Execute)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYouCannotSignYourLpaYetWithUnder18ActorsWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{DateOfBirth: date.New(date.Today().YearString(), "1", "2")},
		}},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, guidanceData{App: testAppData, Donor: donor}).
		Return(expectedError)

	err := YouCannotSignYourLpaYet(template.Execute)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetYouCannotSignYourLpaYetWithoutUnder18Actors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := YouCannotSignYourLpaYet(nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{DateOfBirth: date.New("2000", "1", "2")},
		}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}
