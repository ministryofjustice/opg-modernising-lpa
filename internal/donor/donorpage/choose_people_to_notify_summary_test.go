package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetChoosePeopleToNotifySummary(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{}}}
	peopleToNotify := donordata.PeopleToNotify{{}}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), donor).
		Return(peopleToNotify, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &choosePeopleToNotifySummaryData{
			App:       testAppData,
			Donor:     donor,
			Options:   donordata.YesNoMaybeValues,
			CanChoose: true,
		}).
		Return(nil)

	err := ChoosePeopleToNotifySummary(template.Execute, service)(testAppData, w, r, donor)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetChoosePeopleToNotifySummaryWhenReuseStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{}}}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(r.Context(), donor).
		Return(nil, expectedError)

	err := ChoosePeopleToNotifySummary(nil, service)(testAppData, w, r, donor)
	assert.ErrorIs(t, err, expectedError)
}

func TestGetChoosePeopleToNotifySummaryWhenNoPeopleToNotify(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := ChoosePeopleToNotifySummary(nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID: "lpa-id",
		Tasks: donordata.Tasks{
			YourDetails:                task.StateCompleted,
			ChooseAttorneys:            task.StateCompleted,
			ChooseReplacementAttorneys: task.StateCompleted,
			WhenCanTheLpaBeUsed:        task.StateCompleted,
			Restrictions:               task.StateCompleted,
			CertificateProvider:        task.StateCompleted,
		},
	})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathDoYouWantToNotifyPeople.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifySummaryAddPersonToNotify(t *testing.T) {
	testcases := map[donordata.YesNoMaybe]string{
		donordata.Yes:   donor.PathEnterPersonToNotify.Format("lpa-id") + "?addAnother=1",
		donordata.Maybe: donor.PathChoosePeopleToNotify.Format("lpa-id") + "?addAnother=1",
		donordata.No:    donor.PathTaskList.Format("lpa-id"),
	}

	for value, redirect := range testcases {
		t.Run(value.String(), func(t *testing.T) {
			f := url.Values{
				"option": {value.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			service := newMockPeopleToNotifyService(t)
			service.EXPECT().
				Reusable(mock.Anything, mock.Anything).
				Return(nil, nil)

			err := ChoosePeopleToNotifySummary(nil, service)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{{UID: actoruid.New()}}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, redirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostChoosePeopleToNotifySummaryNoFurtherPeopleToNotify(t *testing.T) {
	f := url.Values{
		"option": {donordata.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(mock.Anything, mock.Anything).
		Return(nil, nil)

	err := ChoosePeopleToNotifySummary(nil, service)(testAppData, w, r, &donordata.Provided{
		LpaID:          "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{{UID: actoruid.New()}},
		Tasks: donordata.Tasks{
			YourDetails:                task.StateCompleted,
			ChooseAttorneys:            task.StateCompleted,
			ChooseReplacementAttorneys: task.StateCompleted,
			WhenCanTheLpaBeUsed:        task.StateCompleted,
			Restrictions:               task.StateCompleted,
			CertificateProvider:        task.StateCompleted,
			PeopleToNotify:             task.StateCompleted,
		},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostChoosePeopleToNotifySummaryFormValidation(t *testing.T) {
	f := url.Values{
		"option": {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With("option", validation.SelectError{Label: "yesToAddAnotherPersonToNotify"})

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Reusable(mock.Anything, mock.Anything).
		Return(nil, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *choosePeopleToNotifySummaryData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := ChoosePeopleToNotifySummary(template.Execute, service)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{{}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
