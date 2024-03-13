package supporter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetViewLPA(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=lpa-id", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id"}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLPAData{
			App:   page.AppData{LpaID: "lpa-id"},
			Donor: donor,
		}).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id", OrganisationID: "org-id", SessionID: "session-id"})).
		Return(donor, nil)

	err := ViewLPA(template.Execute, donorStore)(testAppData, w, r, &actor.Organisation{})

	assert.Nil(t, err)
}
