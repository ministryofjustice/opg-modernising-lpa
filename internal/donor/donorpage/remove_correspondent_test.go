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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRemoveCorrespondent(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	correspondent := donordata.Correspondent{
		UID:        uid,
		FirstNames: "John",
		LastName:   "Smith",
	}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &removeCorrespondentData{
			App:  testAppData,
			Name: "John Smith",
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RemoveCorrespondent(template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{Correspondent: correspondent})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostRemoveCorrespondent(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	correspondent := donordata.Correspondent{
		UID:        uid,
		FirstNames: "John",
		LastName:   "Smith",
	}

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(r.Context(), correspondent).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(r.Context(), event.CorrespondentUpdated{
			UID: "lpa-uid",
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), &donordata.Provided{LpaID: "lpa-id", LpaUID: "lpa-uid"}).
		Return(nil)

	err := RemoveCorrespondent(nil, donorStore, reuseStore, eventClient)(testAppData, w, r, &donordata.Provided{
		LpaID:         "lpa-id",
		LpaUID:        "lpa-uid",
		Correspondent: correspondent,
	})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathAddCorrespondent.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveCorrespondentWhenReuseClientErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCorrespondent(nil, nil, reuseStore, nil)(testAppData, w, r, &donordata.Provided{})
	assert.ErrorIs(t, err, expectedError)
}

func TestPostRemoveCorrespondentWhenEventClientErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCorrespondent(nil, nil, reuseStore, eventClient)(testAppData, w, r, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestPostRemoveCorrespondentWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := RemoveCorrespondent(nil, nil, nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathCorrespondentSummary.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveCorrespondentErrorOnPutStore(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.Yes.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	reuseStore := newMockReuseStore(t)
	reuseStore.EXPECT().
		DeleteCorrespondent(mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCorrespondentUpdated(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCorrespondent(nil, donorStore, reuseStore, eventClient)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.ErrorIs(t, err, expectedError)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRemoveCorrespondentFormValidation(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {""},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	validationError := validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "yesToRemoveCorrespondent"})

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *removeCorrespondentData) bool {
			return assert.Equal(t, validationError, data.Errors)
		})).
		Return(nil)

	err := RemoveCorrespondent(template.Execute, nil, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
