package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourCertificateProviderIsNotRelated(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{
		Tasks: donordata.Tasks{
			CertificateProvider: task.StateCompleted,
		},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &confirmYourCertificateProviderIsNotRelatedData{
			App:   testAppData,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
			Donor: donor,
		}).
		Return(nil)

	err := ConfirmYourCertificateProviderIsNotRelated(template.Execute, nil, nil)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmYourCertificateProviderIsNotRelatedWhenNoCertificateProvider(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ConfirmYourCertificateProviderIsNotRelated(nil, nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestGetConfirmYourCertificateProviderIsNotRelatedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourCertificateProviderIsNotRelated(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{
		Tasks: donordata.Tasks{
			CertificateProvider: task.StateCompleted,
		},
	})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostConfirmYourCertificateProviderIsNotRelated(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:                          "lpa-id",
			Donor:                          donordata.Donor{CanSign: form.Yes},
			HasSentApplicationUpdatedEvent: true,
			Tasks: donordata.Tasks{
				YourDetails:                task.StateCompleted,
				ChooseAttorneys:            task.StateCompleted,
				ChooseReplacementAttorneys: task.StateCompleted,
				WhenCanTheLpaBeUsed:        task.StateCompleted,
				Restrictions:               task.StateCompleted,
				CertificateProvider:        task.StateCompleted,
				PeopleToNotify:             task.StateCompleted,
				AddCorrespondent:           task.StateCompleted,
				CheckYourLpa:               task.StateInProgress,
			},
			CertificateProviderNotRelatedConfirmedAt: testNow,
		}).
		Return(nil)

	err := ConfirmYourCertificateProviderIsNotRelated(nil, donorStore, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID:                          "lpa-id",
		Donor:                          donordata.Donor{CanSign: form.Yes},
		HasSentApplicationUpdatedEvent: true,
		Tasks: donordata.Tasks{
			YourDetails:                task.StateCompleted,
			ChooseAttorneys:            task.StateCompleted,
			ChooseReplacementAttorneys: task.StateCompleted,
			WhenCanTheLpaBeUsed:        task.StateCompleted,
			Restrictions:               task.StateCompleted,
			CertificateProvider:        task.StateCompleted,
			PeopleToNotify:             task.StateCompleted,
			AddCorrespondent:           task.StateCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CheckYourLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourCertificateProviderIsNotRelatedChooseNew(t *testing.T) {
	f := url.Values{
		"action": {"choose-new"},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID: "lpa-id",
			Tasks: donordata.Tasks{
				CertificateProvider: task.StateNotStarted,
			},
		}).
		Return(nil)

	err := ConfirmYourCertificateProviderIsNotRelated(nil, donorStore, testNowFn)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		CertificateProvider: donordata.CertificateProvider{
			FirstNames: "John",
		},
		Tasks: donordata.Tasks{
			CertificateProvider: task.StateCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.CertificateProviderDetails.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourCertificateProviderIsNotRelatedWhenStoreErrors(t *testing.T) {
	testcases := map[string]url.Values{
		"choose-new": {
			"action": {"choose-new"},
		},
		"submit": {
			form.FieldNames.YesNo: {form.Yes.String()},
		},
	}

	for name, form := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), mock.Anything).
				Return(expectedError)

			err := ConfirmYourCertificateProviderIsNotRelated(nil, donorStore, testNowFn)(testAppData, w, r, &donordata.Provided{
				Tasks: donordata.Tasks{
					CertificateProvider: task.StateCompleted,
				},
			})

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestPostConfirmYourCertificateProviderIsNotRelatedWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *confirmYourCertificateProviderIsNotRelatedData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "theBoxToConfirmYourCertificateProviderIsNotRelated"}), data.Errors)
		})).
		Return(nil)

	err := ConfirmYourCertificateProviderIsNotRelated(template.Execute, nil, testNowFn)(testAppData, w, r, &donordata.Provided{
		Tasks: donordata.Tasks{
			CertificateProvider: task.StateCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
