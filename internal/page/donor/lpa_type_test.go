package donor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetLpaType(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaTypeData{
			App:     testAppData,
			Form:    &lpaTypeForm{},
			Options: actor.LpaTypeValues,
		}).
		Return(nil)

	err := LpaType(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaTypeFromStore(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &lpaTypeData{
			App: testAppData,
			Form: &lpaTypeForm{
				LpaType: actor.LpaTypePropertyAndAffairs,
			},
			Options: actor.LpaTypeValues,
		}).
		Return(nil)

	err := LpaType(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{Type: actor.LpaTypePropertyAndAffairs})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetLpaTypeWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := LpaType(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostLpaType(t *testing.T) {
	testcases := map[actor.LpaType]*actor.DonorProvidedDetails{
		actor.LpaTypePropertyAndAffairs: {
			LpaID: "lpa-id",
			Donor: actor.Donor{
				FirstNames:  "John",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "01", "01"),
				Address:     place.Address{Postcode: "F1 1FF"},
			},
			Type:  actor.LpaTypePropertyAndAffairs,
			Tasks: actor.DonorTasks{YourDetails: actor.TaskCompleted},
		},
		actor.LpaTypePersonalWelfare: {
			LpaID: "lpa-id",
			Donor: actor.Donor{
				FirstNames:  "John",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "01", "01"),
				Address:     place.Address{Postcode: "F1 1FF"},
			},
			Type:                actor.LpaTypePersonalWelfare,
			WhenCanTheLpaBeUsed: actor.CanBeUsedWhenCapacityLost,
			Tasks:               actor.DonorTasks{YourDetails: actor.TaskCompleted},
		},
	}

	for lpaType, donor := range testcases {
		t.Run(lpaType.String(), func(t *testing.T) {
			form := url.Values{
				"lpa-type": {lpaType.String()},
			}

			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

			w := httptest.NewRecorder()
			r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(form.Encode()))
			r.Header.Add("Content-Type", page.FormUrlEncoded)

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				Put(r.Context(), donor).
				Return(nil)

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendUidRequested(r.Context(), event.UidRequested{
					LpaID:          "lpa-id",
					DonorSessionID: "an-id",
					Type:           lpaType.String(),
					Donor: uid.DonorDetails{
						Name:     "John Smith",
						Dob:      date.New("2000", "01", "01"),
						Postcode: "F1 1FF",
					},
				}).
				Return(nil)

			err := LpaType(nil, donorStore, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
				LpaID: "lpa-id",
				Donor: actor.Donor{
					FirstNames:  "John",
					LastName:    "Smith",
					DateOfBirth: date.New("2000", "01", "01"),
					Address:     place.Address{Postcode: "F1 1FF"},
				},
				HasSentApplicationUpdatedEvent: true,
			})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, page.Paths.TaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestPostLpaTypeWhenSessionErrors(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	err := LpaType(nil, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
	})

	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestPostLpaTypeWhenEventErrors(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendUidRequested(r.Context(), mock.Anything).
		Return(expectedError)

	err := LpaType(nil, donorStore, eventClient)(testAppData, w, r, &actor.DonorProvidedDetails{
		LpaID: "lpa-id",
	})

	assert.Equal(t, expectedError, err)
}

func TestPostLpaTypeWhenStoreErrors(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		Put(r.Context(), mock.Anything).
		Return(expectedError)

	err := LpaType(nil, donorStore, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})

	assert.Equal(t, expectedError, err)
}

func TestPostLpaTypeWhenValidationErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.MatchedBy(func(data *lpaTypeData) bool {
			return assert.Equal(t, validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}), data.Errors)
		})).
		Return(nil)

	err := LpaType(template.Execute, nil, nil)(testAppData, w, r, &actor.DonorProvidedDetails{})
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReadLpaTypeForm(t *testing.T) {
	form := url.Values{
		"lpa-type": {actor.LpaTypePropertyAndAffairs.String()},
	}

	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", page.FormUrlEncoded)

	result := readLpaTypeForm(r)

	assert.Equal(t, actor.LpaTypePropertyAndAffairs, result.LpaType)
}

func TestLpaTypeFormValidate(t *testing.T) {
	testCases := map[string]struct {
		form   *lpaTypeForm
		errors validation.List
	}{
		"valid": {
			form: &lpaTypeForm{},
		},
		"invalid": {
			form: &lpaTypeForm{
				Error: expectedError,
			},
			errors: validation.With("lpa-type", validation.SelectError{Label: "theTypeOfLpaToMake"}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.errors, tc.form.Validate())
		})
	}
}
