package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRandom struct {
	mock.Mock
}

func (m *mockRandom) String(length int) string {
	args := m.Called(length)
	return args.Get(0).(string)
}

func TestPaymentConfirmation(t *testing.T) {
	t.Run("Gets payment status from GOV UK Pay by payment_id in cookie and stores payment_id and a UUID against users session ID", func(t *testing.T) {
		payClient := &mockPayClient{BaseURL: "http://base.url"}

		payClient.
			On("GetPayment", "abc123").
			Return(pay.GetPaymentResponse{
				State: pay.State{
					Status:   "success",
					Finished: true,
				},
				PaymentId: "abc123",
			}, nil)

		w := httptest.NewRecorder()

		template := &mockTemplate{}
		template.
			On("Func", w, &paymentConfirmationData{App: appData}).
			Return(nil)

		r, _ := http.NewRequest(http.MethodGet, "/payment-confirmation", nil)

		sessionsStore := &mockSessionsStore{}
		getSession := sessions.NewSession(sessionsStore, "pay")

		getSession.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   5400,
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
			Secure:   true,
		}
		getSession.Values = map[interface{}]interface{}{"paymentId": "abc123"}

		sessionsStore.
			On("Get", r, "pay").
			Return(getSession, nil)

		storeSession := sessions.NewSession(sessionsStore, "pay")

		// Expire cookie
		storeSession.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   -1,
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
			Secure:   true,
		}
		storeSession.Values = map[interface{}]interface{}{"paymentId": ""}
		sessionsStore.
			On("Save", r, w, storeSession).
			Return(nil)

		dataStore := &mockDataStore{}
		dataStore.
			On("Get", mock.Anything, "session-id").
			Return(nil)

		dataStore.
			On("Put", mock.Anything, "session-id", Lpa{
				PaymentDetails: PaymentDetails{
					PaymentId:        "abc123",
					PaymentReference: "123456789012",
				},
			}).
			Return(nil)

		random := &mockRandom{}
		random.
			On("String", mock.Anything).
			Return("123456789012")

		err := PaymentConfirmation(&mockLogger{}, template.Func, payClient, dataStore, sessionsStore, random)(appData, w, r)
		resp := w.Result()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mock.AssertExpectationsForObjects(t, template, payClient, dataStore, sessionsStore, random)
	})
}
