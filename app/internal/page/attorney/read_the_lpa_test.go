package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestReadTheLpaWithAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App:      testAppData,
			Lpa:      &page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}},
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadTheLpaWithReplacementAttorney(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{ReplacementAttorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App:      testReplacementAppData,
			Lpa:      &page.Lpa{ReplacementAttorneys: []actor.Attorney{{ID: "attorney-id"}}},
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, lpaStore)(testReplacementAppData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadTheLpaWithAttorneyWhenLpaStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, expectedError)

	err := ReadTheLpa(nil, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadTheLpaWhenAttorneyNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	err := ReadTheLpa(nil, lpaStore)(page.AppData{AttorneyID: "the-wrong-id", ActorType: actor.TypeReplacementAttorney}, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.Paths.Attorney.Start, resp.Header.Get("Location"))
}

func TestReadTheLpaWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStore := newMockLpaStore(t)
	lpaStore.
		On("Get", r.Context()).
		Return(&page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &readTheLpaData{
			App:      testAppData,
			Lpa:      &page.Lpa{Attorneys: []actor.Attorney{{ID: "attorney-id"}}},
			Attorney: actor.Attorney{ID: "attorney-id"},
		}).
		Return(expectedError)

	err := ReadTheLpa(template.Execute, lpaStore)(testAppData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
