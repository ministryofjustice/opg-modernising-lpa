package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/forms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPhoneNumber(t *testing.T) {
	testcases := map[string]struct {
		appData appcontext.Data
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &phoneNumberData{
					App:  tc.appData,
					Form: newPhoneNumberForm(tc.appData.Localizer),
				}).
				Return(nil)

			err := PhoneNumber(template.Execute, nil)(tc.appData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPhoneNumberFromStore(t *testing.T) {
	testcases := map[string]struct {
		appData  appcontext.Data
		attorney *attorneydata.Provided
	}{
		"attorney": {
			appData: testAppData,
		},
		"replacement attorney": {
			appData: testReplacementAppData,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			form := newPhoneNumberForm(tc.appData.Localizer)
			form.Phone.Set("07535111222")

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &phoneNumberData{
					App:  tc.appData,
					Form: form,
				}).
				Return(nil)

			err := PhoneNumber(template.Execute, nil)(tc.appData, w, r, &attorneydata.Provided{
				Phone:    "07535111222",
				PhoneSet: true,
			}, &lpadata.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetPhoneNumberWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := PhoneNumber(template.Execute, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPhoneNumber(t *testing.T) {
	testCases := map[string]struct {
		form            url.Values
		attorney        *attorneydata.Provided
		updatedAttorney *attorneydata.Provided
		appData         appcontext.Data
	}{
		"attorney": {
			form: url.Values{
				"phone": {"07535111222"},
			},
			attorney: &attorneydata.Provided{LpaID: "lpa-id"},
			updatedAttorney: &attorneydata.Provided{
				LpaID:    "lpa-id",
				Phone:    "07535111222",
				PhoneSet: true,
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
			appData: testAppData,
		},
		"attorney empty": {
			appData: testAppData,
			attorney: &attorneydata.Provided{
				LpaID: "lpa-id",
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
			updatedAttorney: &attorneydata.Provided{
				LpaID:    "lpa-id",
				PhoneSet: true,
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
				},
			},
		},
		"replacement attorney": {
			form: url.Values{
				"phone": {"07535111222"},
			},
			attorney: &attorneydata.Provided{LpaID: "lpa-id"},
			updatedAttorney: &attorneydata.Provided{
				LpaID:    "lpa-id",
				Phone:    "07535111222",
				PhoneSet: true,
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
			appData: testReplacementAppData,
		},
		"replacement attorney empty": {
			appData:  testReplacementAppData,
			attorney: &attorneydata.Provided{LpaID: "lpa-id"},
			updatedAttorney: &attorneydata.Provided{
				LpaID:    "lpa-id",
				PhoneSet: true,
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateInProgress,
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(tc.form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				Put(r.Context(), tc.updatedAttorney).
				Return(nil)

			err := PhoneNumber(nil, attorneyStore)(tc.appData, w, r, tc.attorney, &lpadata.Lpa{})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, attorney.PathYourPreferredLanguage.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostPhoneNumberWhenValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	form := url.Values{
		"phone": {"abcd"},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *phoneNumberData) bool {
			return assert.Equal(t, []forms.Field{data.Form.Phone.Field}, data.Form.Errors) &&
				assert.Equal(t, "errorPhone:Label=phone", data.Form.Phone.Error.Format(testAppData.Localizer))
		})).
		Return(nil)

	err := PhoneNumber(template.Execute, nil)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostPhoneNumberWhenAttorneyStoreErrors(t *testing.T) {
	form := url.Values{
		"phone": {"07535111222"},
	}

	w := httptest.NewRecorder()

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := PhoneNumber(nil, attorneyStore)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
