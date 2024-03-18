package donor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

var mockUuidString = func() string { return "123" }

var (
	testAddress = place.Address{
		Line1:      "a",
		Line2:      "b",
		Line3:      "c",
		TownOrCity: "d",
		Postcode:   "E",
		Country:    "GB",
	}
	expectedError = errors.New("err")
	testAppData   = page.AppData{
		SessionID:         "session-id",
		LpaID:             "lpa-id",
		Lang:              localize.En,
		LoginSessionEmail: "logged-in@example.com",
	}
	testSupporterAppData = page.AppData{
		SessionID:   "session-id",
		LpaID:       "lpa-id",
		IsSupporter: true,
		Lang:        localize.En,
	}
	testNow   = time.Date(2023, time.July, 3, 4, 5, 6, 1, time.UTC)
	testNowFn = func() time.Time { return testNow }
	testUID   = actoruid.New()
	testUIDFn = func() actoruid.UID { return testUID }
)

func (m *mockDonorStore) withCompletedPaymentLpaData(r *http.Request, paymentId, paymentReference string, paymentAmount int) *mockDonorStore {
	m.EXPECT().
		Put(r.Context(), &actor.DonorProvidedDetails{
			CertificateProvider: actor.CertificateProvider{
				Email: "certificateprovider@example.com",
			},
			PaymentDetails: []actor.Payment{{
				PaymentId:        paymentId,
				PaymentReference: paymentReference,
				Amount:           paymentAmount,
			}},
			Tasks: actor.DonorTasks{
				PayForLpa: actor.PaymentTaskCompleted,
			},
		}).
		Return(nil)

	return m
}

func (m *mockSessionStore) withPaySession(r *http.Request) *mockSessionStore {
	m.EXPECT().Payment(r).Return(&sesh.PaymentSession{PaymentID: "abc123"}, nil)

	return m
}

func (m *mockSessionStore) withExpiredPaySession(r *http.Request, w *httptest.ResponseRecorder) *mockSessionStore {
	m.EXPECT().ClearPayment(r, w).Return(nil)

	return m
}
