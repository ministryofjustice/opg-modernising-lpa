package supporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGetViewLPA(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLPAData{
			App:   testAppData,
			Donor: donor,
		}).
		Return(nil)

	err := ViewLPA(template.Execute, donorStore)(testAppData, w, r, &actor.Organisation{})

	assert.Nil(t, err)
}

func TestGetViewLPAWithSessionMissing(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?id=lpa-id", nil)

	err := ViewLPA(nil, nil)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAWithDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(nil, expectedError)

	err := ViewLPA(nil, donorStore)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}

func TestGetViewLPAWhenTemplateError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(page.ContextWithSessionData(context.Background(), &page.SessionData{}), http.MethodGet, "/?id=lpa-id", nil)

	donor := &actor.DonorProvidedDetails{LpaID: "lpa-id"}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Get(page.ContextWithSessionData(r.Context(), &page.SessionData{LpaID: "lpa-id"})).
		Return(donor, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &viewLPAData{
			App:   testAppData,
			Donor: donor,
		}).
		Return(expectedError)

	err := ViewLPA(template.Execute, donorStore)(testAppData, w, r, &actor.Organisation{})

	assert.Error(t, err)
}
