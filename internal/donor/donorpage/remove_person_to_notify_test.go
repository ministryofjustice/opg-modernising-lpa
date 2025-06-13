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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemovePersonToNotify(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uid.String(), nil)

	person := donordata.PersonToNotify{
		UID: uid,
		Address: place.Address{
			Line1: "1 Road way",
		},
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removePersonToNotifyData{
			App:            testAppData,
			PersonToNotify: person,
			Errors:         nil,
			Form:           form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemovePersonToNotify(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		PeopleToNotify: donordata.PeopleToNotify{person},
	})

	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRemovePersonToNotifyWhenDoesNotExist(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=invalid-id", nil)

	err := RemovePersonToNotify(nil, nil)(testAppData, w, r, &donordata.Provided{
		LpaID:          "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{{UID: testUID}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotify(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	person := donordata.PersonToNotify{UID: testUID}

	provided := &donordata.Provided{
		LpaID:          "lpa-id",
		PeopleToNotify: donordata.PeopleToNotify{person},
	}

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Delete(r.Context(), provided, person).
		Return(nil)

	err := RemovePersonToNotify(nil, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotifyWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := RemovePersonToNotify(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id", PeopleToNotify: donordata.PeopleToNotify{{UID: testUID}}})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathChoosePeopleToNotifySummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemovePersonToNotifyWhenServiceErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	service := newMockPeopleToNotifyService(t)
	service.EXPECT().
		Delete(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemovePersonToNotify(nil, service)(testAppData, w, r, &donordata.Provided{
		PeopleToNotify: donordata.PeopleToNotify{{UID: testUID}},
	})
	assert.ErrorIs(t, err, expectedError)
}

func TestRemovePersonToNotifyFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveThisPerson"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removePersonToNotifyData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemovePersonToNotify(template.Execute, nil)(testAppData, w, r, &donordata.Provided{
		PeopleToNotify: donordata.PeopleToNotify{{UID: testUID}},
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
