package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetReadTheLpaWithAttorney(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &readTheLpaData{
			App:   testAppData,
			Donor: &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, donorStore, nil)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWithReplacementAttorney(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &readTheLpaData{
			App:   testReplacementAppData,
			Donor: &actor.DonorProvidedDetails{ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}},
		}).
		Return(nil)

	err := ReadTheLpa(template.Execute, donorStore, nil)(testReplacementAppData, w, r, &actor.AttorneyProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWithAttorneyWhenDonorStoreErrors(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}}, expectedError)

	err := ReadTheLpa(nil, donorStore, nil)(testAppData, w, r, nil)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetReadTheLpaWhenTemplateError(t *testing.T) {
	uid := actoruid.New()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(r.Context()).
		Return(&actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &readTheLpaData{
			App:   testAppData,
			Donor: &actor.DonorProvidedDetails{Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}}},
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
	attorneyStore.EXPECT().
		Put(r.Context(), &actor.AttorneyProvidedDetails{
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
