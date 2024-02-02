package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	actor "github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrganisationCreated(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, organisationCreatedData{App: testAppData, OrganisationName: "A name"}).
		Return(nil)

	err := OrganisationCreated(template.Execute)(testAppData, w, r, &actor.Organisation{Name: "A name"})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestOrganisationCreatedWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := OrganisationCreated(template.Execute)(testAppData, w, r, &actor.Organisation{Name: "A name"})
	assert.Equal(t, expectedError, err)
}
