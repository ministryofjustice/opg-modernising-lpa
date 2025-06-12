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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemovePersonToNotify(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	personToNotify := donordata.PersonToNotify{
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

	err := RemovePersonToNotify(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotify}})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemovePersonToNotifyAttorneyDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	personToNotify := donordata.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	err := RemovePersonToNotify(nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{personToNotify}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotify(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := donordata.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := donordata.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeletePersonToNotify(r.Context(), personToNotifyWithoutAddress).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithAddress}}).
		Return(nil)

	err := RemovePersonToNotify(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotifyWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := donordata.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := donordata.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	err := RemovePersonToNotify(nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotifyWhenReuseStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := donordata.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := donordata.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeletePersonToNotify(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemovePersonToNotify(nil, nil, reuseStore)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostRemovePersonToNotifyWhenDonorStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithAddress := donordata.PersonToNotify{
		UID: actoruid.New(),
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	personToNotifyWithoutAddress := donordata.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeletePersonToNotify(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemovePersonToNotify(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithoutAddress, personToNotifyWithAddress}})
	assert.ErrorIs(t, err, expectedError)
}

func TestRemovePersonToNotifyFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+uid.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	personToNotifyWithoutAddress := donordata.PersonToNotify{
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

	err := RemovePersonToNotify(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithoutAddress}})
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

	personToNotifyWithoutAddress := donordata.PersonToNotify{
		UID:     uid,
		Address: place.Address{},
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeletePersonToNotify(r.Context(), personToNotifyWithoutAddress).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{
			LpaID:          "lpa-id",
			PeopleToNotify: donordata.PeopleToNotify{},
			Tasks:          donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted, PeopleToNotify: task.StateNotStarted},
		}).
		Return(nil)

	err := RemovePersonToNotify(nil, donorStore, reuseStore)(testAppData, w, r, &donordata.Provided{
		LpaID:          "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{personToNotifyWithoutAddress},
		Tasks:          donordata.Tasks{YourDetails: task.StateCompleted, ChooseAttorneys: task.StateCompleted, PeopleToNotify: task.StateCompleted},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}
