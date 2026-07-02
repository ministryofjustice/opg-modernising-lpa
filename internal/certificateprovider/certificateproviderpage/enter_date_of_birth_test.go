package certificateproviderpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEnterDateOfBirth(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &lpadata.Lpa{
		LpaID: "lpa-id",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dateOfBirthData{
			App:  testAppData,
			Lpa:  lpa,
			Form: newDateOfBirthForm(testAppData.Localizer),
		}).
		Return(nil)

	err := EnterDateOfBirth(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, lpa)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterDateOfBirthFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	form := newDateOfBirthForm(testAppData.Localizer)
	form.Dob.Set(date.New("1997", "1", "2"))

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dateOfBirthData{
			App:  testAppData,
			Lpa:  &lpadata.Lpa{},
			Form: form,
		}).
		Return(nil)

	err := EnterDateOfBirth(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{DateOfBirth: date.New("1997", "1", "2")}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEnterDateOfBirthWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpadata.Lpa{
		LpaID: "lpa-id",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &dateOfBirthData{
			App:  testAppData,
			Lpa:  donor,
			Form: newDateOfBirthForm(testAppData.Localizer),
		}).
		Return(expectedError)

	err := EnterDateOfBirth(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, donor)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostEnterDateOfBirth(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form      url.Values
		retrieved *certificateproviderdata.Provided
		updated   *certificateproviderdata.Provided
	}{
		"valid": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			retrieved: &certificateproviderdata.Provided{LpaID: "lpa-id"},
			updated: &certificateproviderdata.Provided{
				LpaID:       "lpa-id",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
		},
		"previously completed": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			retrieved: &certificateproviderdata.Provided{
				LpaID: "lpa-id",
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			updated: &certificateproviderdata.Provided{
				LpaID:       "lpa-id",
				DateOfBirth: date.New(validBirthYear, "1", "2"),
				Tasks: certificateproviderdata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				Put(r.Context(), tc.updated).
				Return(nil)

			err := EnterDateOfBirth(nil, certificateProviderStore)(testAppData, w, r, tc.retrieved, &lpadata.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, certificateprovider.PathYourPreferredLanguage.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostEnterDateOfBirthWhenInputRequired(t *testing.T) {
	validBirthYear := strconv.Itoa(time.Now().Year() - 40)

	testCases := map[string]struct {
		form        url.Values
		dataMatcher func(t *testing.T, data *dateOfBirthData) bool
	}{
		"validation error": {
			form: url.Values{
				"date-of-birth-day":   {"55"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {validBirthYear},
			},
			dataMatcher: func(t *testing.T, data *dateOfBirthData) bool {
				return assert.Equal(t, []forms.Field{data.Form.Dob.Field}, data.Form.Errors) &&
					assert.Equal(t, "errorDateMustBeReal:Label=dateOfBirth", data.Form.Dob.Error.Format(testAppData.Localizer))
			},
		},
		"dob warning": {
			form: url.Values{
				"date-of-birth-day":   {"2"},
				"date-of-birth-month": {"1"},
				"date-of-birth-year":  {"1900"},
			},
			dataMatcher: func(t *testing.T, data *dateOfBirthData) bool {
				return assert.Equal(t, "dateOfBirthIsOver100", data.DobWarning)
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, mock.MatchedBy(func(data *dateOfBirthData) bool {
					return tc.dataMatcher(t, data)
				})).
				Return(nil)

			err := EnterDateOfBirth(template.Execute, nil)(testAppData, w, r, &certificateproviderdata.Provided{}, &lpadata.Lpa{LpaID: "lpa-id"})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
