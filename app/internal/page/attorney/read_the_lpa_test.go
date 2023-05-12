package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetReadTheLpaWithAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App:      testAppData,
			Lpa:      &page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}},
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, lpaStore, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWithReplacementAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App:      testReplacementAppData,
			Lpa:      &page.Lpa{ReplacementAttorneys: []actor.Attorney{{ID: "attorney-id"}}},
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, lpaStore, attorneyStore)(testReplacementAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWithAttorneyWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, expectedError)

	err := ReadTheLpa(nil, lpaStore, nil)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWhenAttorneyNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	err := ReadTheLpa(nil, lpaStore, nil)(page.AppData{AttorneyID: "the-wrong-id", ActorType: actor.TypeReplacementAttorney}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start, resp.Header.Get("Location"))
}

func TestGetReadTheLpaWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App:      testAppData,
			Lpa:      &page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}},
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, lpaStore, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostReadTheLpaWithAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)
	attorneyStore.
		On("Put", r.Context(), &actor.AttorneyProvidedDetails{
			Tasks: actor.AttorneyTasks{
				ReadTheLpa: actor.TaskCompleted,
			},
		}).
		Return(nil)

	err := ReadTheLpa(nil, lpaStore, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.RightsAndResponsibilities, resp.Header.Get("Location"))
}

func TestPostReadTheLpaWithAttorneyOnLpaStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("Get", r.Context()).
		Return(&actor.AttorneyProvidedDetails{}, nil)
	attorneyStore.
		On("Put", r.Context(), &actor.AttorneyProvidedDetails{
			Tasks: actor.AttorneyTasks{
				ReadTheLpa: actor.TaskCompleted,
			},
		}).
		Return(expectedError)

	err := ReadTheLpa(nil, lpaStore, attorneyStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
