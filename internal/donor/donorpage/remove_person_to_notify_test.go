package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemovePersonToNotify(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	personToNotify := actor.PersonToNotify{
		UID: uid,
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removePersonToNotifyData{
			App:            testAppData,
			PersonToNotify: personToNotify,
			Errors:         nil,
			Form:           form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemovePersonToNotify(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{PeopleToNotify: actor.PeopleToNotify{personToNotify}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemovePersonToNotifyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	personToNotify := actor.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	err := RemovePersonToNotify(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", PeopleToNotify: actor.PeopleToNotify{personToNotify}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotify(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := actor.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := actor.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{LpaID: "lpa-id", PeopleToNotify: actor.PeopleToNotify{personToNotifyWithAddress}}).
		Return(nil)

	err := RemovePersonToNotify(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", PeopleToNotify: actor.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotifyWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := actor.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := actor.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	err := RemovePersonToNotify(nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{LpaID: "lpa-id", PeopleToNotify: actor.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotifyErrorOnPutStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := actor.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := actor.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{PeopleToNotify: actor.PeopleToNotify{personToNotifyWithAddress}}).
		Return(expectedError)

	err := RemovePersonToNotify(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{PeopleToNotify: actor.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemovePersonToNotifyFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithoutAddress := actor.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveThisPerson"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removePersonToNotifyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemovePersonToNotify(template.Execute, nil)(testAppData, w, r, &actor.DonorProvidedDetails{PeopleToNotify: actor.PeopleToNotify{personToNotifyWithoutAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemovePersonToNotifyRemoveLastPerson(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithoutAddress := actor.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			LpaID:          "lpa-id",
			PeopleToNotify: actor.PeopleToNotify{},
			Tasks:          actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, PeopleToNotify: actor.TaskNotStarted},
		}).
		Return(nil)

	err := RemovePersonToNotify(nil, donorStore)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID:          "lpa-id",
		PeopleToNotify: actor.PeopleToNotify{personToNotifyWithoutAddress},
		Tasks:          actor.DonorTasks{YourDetails: actor.TaskCompleted, ChooseAttorneys: actor.TaskCompleted, PeopleToNotify: actor.TaskCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.ChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}
