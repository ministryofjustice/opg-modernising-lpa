package donorpage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRegisterWithCourtOfProtection(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &registerWithCourtOfProtectionData{
			App:  testAppData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}).
		Return(nil)

	err := RegisterWithCourtOfProtection(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetRegisterWithCourtOfProtectionWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := RegisterWithCourtOfProtection(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostRegisterWithCourtOfProtection(t *testing.T) {
	testCases := map[string]struct {
		yesNo            form.YesNo
		donorStore       func() *mockDonorStore
		eventClient      func() *mockEventClient
		expectedRedirect string
	}{
		"yes": {
			yesNo:            form.Yes,
			expectedRedirect: donor.PathDeleteThisLpa.Format("lpa-id"),
			donorStore:       func() *mockDonorStore { return nil },
			eventClient:      func() *mockEventClient { return nil },
		},
		"no": {
			yesNo: form.No,
			donorStore: func() *mockDonorStore {
				donorStore := newMockDonorStore(t)
				donorStore.EXPECT().
					Put(context.Background(), &donordata.Provided{LpaID: "lpa-id", LpaUID: "lpa-uid", RegisteringWithCourtOfProtection: true}).
					Return(nil)
				return donorStore
			},
			eventClient: func() *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendRegisterWithCourtOfProtection(context.Background(), event.RegisterWithCourtOfProtection{
						UID: "lpa-uid",
					}).
					Return(nil)
				return eventClient
			},
			expectedRedirect: donor.PathWhatHappensNextRegisteringWithCourtOfProtection.Format("lpa-id"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			f := url.Values{
				form.FieldNames.YesNo: {tc.yesNo.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			err := RegisterWithCourtOfProtection(nil, tc.donorStore(), tc.eventClient())(testAppData, w, r, &donordata.Provided{
				LpaID:  "lpa-id",
				LpaUID: "lpa-uid",
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, tc.expectedRedirect, resp.Header.Get("Location"))
		})
	}
}

func TestPostRegisterWithCourtOfProtectionWhenEventClientErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendRegisterWithCourtOfProtection(mock.Anything, mock.Anything).
		Return(expectedError)

	err := RegisterWithCourtOfProtection(nil, nil, eventClient)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostRegisterWithCourtOfProtectionWhenStoreErrors(t *testing.T) {
	f := url.Values{
		form.FieldNames.YesNo: {form.No.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(f.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendRegisterWithCourtOfProtection(mock.Anything, mock.Anything).
		Return(nil)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := RegisterWithCourtOfProtection(nil, donorStore, eventClient)(testAppData, w, r, &donordata.Provided{})

	assert.Equal(t, expectedError, err)
}

func TestPostRegisterWithCourtOfProtectionWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *registerWithCourtOfProtectionData) bool {
			return assert.Equal(t, validation.With(form.FieldNames.YesNo, validation.SelectError{Label: "whatYouWouldLikeToDo"}), data.Errors)
		})).
		Return(nil)

	err := RegisterWithCourtOfProtection(template.Execute, nil, nil)(testAppData, w, r, &donordata.Provided{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
