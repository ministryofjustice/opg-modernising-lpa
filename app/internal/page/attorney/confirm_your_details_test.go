package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetConfirmYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	attorneyProvidedDetails := &actor.AttorneyProvidedDetails{ID: "123"}

	attorney := actor.Attorney{
		ID:         "123",
		FirstNames: "John",
	}

	lpa := &page.Lpa{
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{attorney}},
	}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &confirmYourDetailsData{
			App:                     testAppData,
			Lpa:                     lpa,
			Attorney:                attorney,
			AttorneyProvidedDetails: attorneyProvidedDetails,
		}).
		Return(nil)

	err := ConfirmYourDetails(template.Execute, nil, donorStore)(testAppData, w, r, attorneyProvidedDetails)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConfirmYourDetailsWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &page.Lpa{}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, expectedError)

	err := ConfirmYourDetails(nil, nil, donorStore)(testAppData, w, r, nil)

	assert.Equal(t, expectedError, err)
}

func TestGetConfirmYourDetailsWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(template.Execute, nil, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostConfirmYourDetails(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Put", r.Context(), &actor.AttorneyProvidedDetails{
			ID:    "123",
			LpaID: "lpa-id",
			Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskCompleted},
		}).
		Return(nil)

	err := ConfirmYourDetails(nil, attorneyStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{ID: "123", LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.ReadTheLpa.Format("lpa-id"), resp.Header.Get("Location"))
}

func TestPostConfirmYourDetailsWhenStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Put", r.Context(), mock.Anything).
		Return(expectedError)

	err := ConfirmYourDetails(nil, attorneyStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{ID: "123", LpaID: "lpa-id"})
	assert.Equal(t, expectedError, err)
}
