package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAddCorrespondent(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &addCorrespondentData{
			App:   testAppData,
			Donor: &donordata.Provided{},
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAddCorrespondentFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &addCorrespondentData{
			App:   testAppData,
			Donor: &donordata.Provided{AddCorrespondent: form.Yes},
			Form:  form.NewYesNoForm(form.Yes),
		}).
		Return(nil)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{AddCorrespondent: form.Yes})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAddCorrespondentWhenExists(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := AddCorrespondent(nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:         "lpa-id",
		Correspondent: donordata.Correspondent{UID: testUID},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCorrespondentSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetAddCorrespondentWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostAddCorrespondentWhenYes(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:            "lpa-id",
			LpaUID:           "lpa-uid",
			AddCorrespondent: form.Yes,
			Correspondent:    donordata.Correspondent{FirstNames: "John"},
		}).
		Return(nil)

	err := AddCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{
		LpaID:         "lpa-id",
		LpaUID:        "lpa-uid",
		Correspondent: donordata.Correspondent{FirstNames: "John"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChooseCorrespondent.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostAddCorrespondentWhenNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockCorrespondentService(t)
	service.EXPECT().
		NotWanted(r.Context(), &donordata.Provided{
			LpaID:            "lpa-id",
			LpaUID:           "lpa-uid",
			AddCorrespondent: form.No,
			Correspondent:    donordata.Correspondent{FirstNames: "John"},
		}).
		Return(nil)

	err := AddCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{
		LpaID:         "lpa-id",
		LpaUID:        "lpa-uid",
		Correspondent: donordata.Correspondent{FirstNames: "John"},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostAddCorrespondentWhenServiceErrors(t *testing.T) {
	testcases := map[form.YesNo]func(*mockCorrespondentService){
		form.Yes: func(service *mockCorrespondentService) {
			service.EXPECT().
				Put(mock.Anything, mock.Anything).
				Return(expectedError)
		},
		form.No: func(service *mockCorrespondentService) {
			service.EXPECT().
				NotWanted(mock.Anything, mock.Anything).
				Return(expectedError)
		},
	}

	for yesNo, setupService := range testcases {
		t.Run(yesNo.String(), func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := newMockCorrespondentService(t)
			setupService(service)

			err := AddCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{})

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestPostAddCorrespondentWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *addCorrespondentData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToAddCorrespondent"}), data.Errors)
		})).
		Return(nil)

	err := AddCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
