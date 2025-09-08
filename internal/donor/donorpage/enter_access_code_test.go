package donorpage

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnterAccessCode(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	session := &sesh.LoginSession{
		Sub:     "hey",
		Email:   "a@example.com",
		HasLPAs: true,
	}

	accessCode := accesscodedata.Link{
		LpaKey:      dynamo.LpaKey("lpa-id"),
		LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("123")),
		ActorUID:    testUID,
		LpaUID:      "lpa-uid",
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(r.Context(), accessCode, "a@example.com").
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(r.Context(), "donor access added", slog.String("lpa_id", "lpa-id"))

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(r.Context(), event.CategoryFunnelStartRate, event.MeasureOnlineDonor).
		Return(nil)

	err := EnterAccessCode(logger, donorStore, eventClient)(testAppData, w, r, session, accessCode)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, page.PathDashboard.Format(), resp.Header.Get("Location"))
}

func TestEnterAccessCodeOnDonorStoreError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(nil, donorStore, nil)(testAppData, w, r, &sesh.LoginSession{}, accesscodedata.Link{
		LpaKey: dynamo.LpaKey(""),
	})
	assert.ErrorIs(t, err, expectedError)
}

func TestEnterAccessCodeOnEventClientError(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Link(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(mock.Anything, mock.Anything, mock.Anything)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendMetric(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := EnterAccessCode(logger, donorStore, eventClient)(testAppData, w, r, &sesh.LoginSession{}, accesscodedata.Link{
		LpaKey: dynamo.LpaKey(""),
	})
	assert.ErrorIs(t, err, expectedError)
}
