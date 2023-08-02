package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetReadTheLpaWithAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App: testAppData,
			Lpa: &page.Lpa{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, donorStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWithReplacementAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App: testReplacementAppData,
			Lpa: &page.Lpa{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, donorStore, nil)(testReplacementAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWithAttorneyWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}}, expectedError)

	err := ReadTheLpa(nil, donorStore, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&page.Lpa{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App: testAppData,
			Lpa: &page.Lpa{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}},
		}).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, donorStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostReadTheLpa(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Put", r.Context(), &actor.AttorneyProvidedDetails{
			LpaID: "lpa-id",
			Tasks: actor.AttorneyTasks{
				ReadTheLpa: actor.TaskCompleted,
			},
		}).
		Return(nil)

	err := ReadTheLpa(nil, nil, attorneyStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.RightsAndResponsibilities.Format("lpa-id"), resp.Header.Get("Location"))
}
