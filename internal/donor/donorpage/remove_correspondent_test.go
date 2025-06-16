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

	err := RemoveCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{Correspondent: correspondent})
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

	provided := &donordata.Provided{
		LpaID:  "lpa-id",
		LpaUID: "lpa-uid",
		Correspondent: donordata.Correspondent{
			UID:        uid,
			FirstNames: "John",
			LastName:   "Smith",
		},
	}

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Delete(r.Context(), provided).
		Return(nil)

	err := RemoveCorrespondent(nil, service)(testAppData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, donor.PathAddCorrespondent.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostRemoveCorrespondentWithFormValueNo(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := RemoveCorrespondent(nil, nil)(testAppData, w, r, &donordata.Provided{LpaID: "lpa-id"})
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

	service := newMockCorrespondentService(t)
	service.EXPECT().
		Delete(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RemoveCorrespondent(nil, service)(testAppData, w, r, &donordata.Provided{})
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

	err := RemoveCorrespondent(template.Execute, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
