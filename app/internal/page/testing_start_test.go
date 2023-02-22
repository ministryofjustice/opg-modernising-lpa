package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestingStart(t *testing.T) {
	t.Run("payment not complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{ID: "123"}).
			Return(nil)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("payment complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&paymentComplete=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:    "123",
				Tasks: Tasks{PayForLpa: TaskCompleted},
			}).
			Return(nil)

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with payment", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&paymentComplete=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:    "123",
				Tasks: Tasks{PayForLpa: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore)
	})

	t.Run("with attorney", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorney=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				Attorneys: actor.Attorneys{
					{
						ID:          "JohnSmith",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				Tasks: Tasks{
					ChooseAttorneys: TaskCompleted,
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with incomplete attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withIncompleteAttorneys=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		attorneys := actor.Attorneys{
			{
				ID:          "with-address",
				FirstNames:  "John",
				LastName:    "Smith",
				Email:       TestEmail,
				DateOfBirth: date.New("2000", "1", "2"),
				Address: place.Address{
					Line1:      "2 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				},
			},
			{
				ID:          "without-address",
				FirstNames:  "Joan",
				LastName:    "Smith",
				Email:       TestEmail,
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{},
			},
		}

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                                   "123",
				Type:                                 LpaTypePropertyFinance,
				WhenCanTheLpaBeUsed:                  UsedWhenRegistered,
				Attorneys:                            attorneys,
				ReplacementAttorneys:                 attorneys,
				HowAttorneysMakeDecisions:            JointlyAndSeverally,
				WantReplacementAttorneys:             "yes",
				HowReplacementAttorneysMakeDecisions: JointlyAndSeverally,
				HowShouldReplacementAttorneysStepIn:  OneCanNoLongerAct,
				Tasks: Tasks{
					ChooseAttorneys:            TaskInProgress,
					ChooseReplacementAttorneys: TaskInProgress,
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorneys=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		attorneys := actor.Attorneys{
			{
				ID:          "JohnSmith",
				FirstNames:  "John",
				LastName:    "Smith",
				Email:       TestEmail,
				DateOfBirth: date.New("2000", "1", "2"),
				Address: place.Address{
					Line1:      "2 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				},
			},
			{
				ID:          "JoanSmith",
				FirstNames:  "Joan",
				LastName:    "Smith",
				Email:       TestEmail,
				DateOfBirth: date.New("2000", "1", "2"),
				Address: place.Address{
					Line1:      "2 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				},
			},
		}

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                        "123",
				Attorneys:                 attorneys,
				HowAttorneysMakeDecisions: JointlyAndSeverally,
				Tasks: Tasks{
					ChooseAttorneys: TaskCompleted,
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("how attorneys act", func(t *testing.T) {
		testCases := []struct {
			DecisionsType    string
			DecisionsDetails string
		}{
			{DecisionsType: "jointly", DecisionsDetails: ""},
			{DecisionsType: "jointly-and-severally", DecisionsDetails: ""},
			{DecisionsType: "mixed", DecisionsDetails: "some details"},
		}

		for _, tc := range testCases {
			t.Run(tc.DecisionsType, func(t *testing.T) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/?redirect=/somewhere&howAttorneysAct=%s", tc.DecisionsType), nil)
				ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

				sessionsStore := &mockSessionsStore{}
				sessionsStore.
					On("Save", r, w, mock.Anything).
					Return(nil)

				lpaStore := newMockLpaStore(t)
				lpaStore.
					On("Create", ctx).
					Return(&Lpa{ID: "123"}, nil)
				lpaStore.
					On("Put", ctx, &Lpa{
						ID:                               "123",
						HowAttorneysMakeDecisions:        tc.DecisionsType,
						HowAttorneysMakeDecisionsDetails: tc.DecisionsDetails,
					}).
					Return(nil)

				TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
				mock.AssertExpectationsForObjects(t, sessionsStore)
			})
		}
	})

	t.Run("with Certificate Provider", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withCP=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				CertificateProvider: actor.CertificateProvider{
					FirstNames:              "Barbara",
					LastName:                "Smith",
					Email:                   TestEmail,
					Mobile:                  TestMobile,
					DateOfBirth:             date.New("1997", "1", "2"),
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
				},
				Tasks: Tasks{CertificateProvider: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with donor details", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withDonorDetails=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				You: actor.Person{
					FirstNames: "Jose",
					LastName:   "Smith",
					Address: place.Address{
						Line1:      "1 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
					Email:       TestEmail,
					DateOfBirth: date.New("2000", "1", "2"),
				},
				WhoFor: "me",
				Type:   LpaTypePropertyFinance,
				Tasks:  Tasks{YourDetails: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with replacement attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withReplacementAttorneys=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                                   "123",
				WantReplacementAttorneys:             "yes",
				HowReplacementAttorneysMakeDecisions: JointlyAndSeverally,
				HowShouldReplacementAttorneysStepIn:  OneCanNoLongerAct,
				Tasks:                                Tasks{ChooseReplacementAttorneys: TaskCompleted},
				ReplacementAttorneys: actor.Attorneys{
					{
						FirstNames: "Jane",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JaneSmith",
					},
					{
						FirstNames: "Jorge",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JorgeSmith",
					},
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("when can be used completed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&whenCanBeUsedComplete=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                  "123",
				WhenCanTheLpaBeUsed: UsedWhenRegistered,
				Tasks:               Tasks{WhenCanTheLpaBeUsed: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with restrictions", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withRestrictions=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:           "123",
				Restrictions: "Some restrictions on how Attorneys act",
				Tasks:        Tasks{Restrictions: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with people to notify", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withPeopleToNotify=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                      "123",
				DoYouWantToNotifyPeople: "yes",
				Tasks:                   Tasks{PeopleToNotify: TaskCompleted},
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      TestEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JonathanSmith",
						FirstNames: "Jonathan",
						LastName:   "Smith",
						Email:      TestEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("with incomplete people to notify", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withIncompletePeopleToNotify=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:                      "123",
				DoYouWantToNotifyPeople: "yes",
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      TestEmail,
						Address:    place.Address{},
					},
				},
				Tasks: Tasks{PeopleToNotify: TaskInProgress},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("lpa checked", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&lpaChecked=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID:           "123",
				Checked:      true,
				HappyToShare: true,
				Tasks:        Tasks{CheckYourLpa: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("id confirmed and signed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&idConfirmedAndSigned=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				OneLoginUserData: identity.UserData{
					OK:          true,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FullName:    "Jose Smith",
				},
				WantToApplyForLpa:      true,
				WantToSignLpa:          true,
				CPWitnessCodeValidated: true,
				Submitted:              time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				Tasks:                  Tasks{ConfirmYourIdentityAndSign: TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("complete LPA", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&completeLpa=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				OneLoginUserData: identity.UserData{
					OK:          true,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FullName:    "Jose Smith",
				},
				WantToApplyForLpa:       true,
				WantToSignLpa:           true,
				CPWitnessCodeValidated:  true,
				Submitted:               time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				Checked:                 true,
				HappyToShare:            true,
				DoYouWantToNotifyPeople: "yes",
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      TestEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JonathanSmith",
						FirstNames: "Jonathan",
						LastName:   "Smith",
						Email:      TestEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				Restrictions:                         "Some restrictions on how Attorneys act",
				WhenCanTheLpaBeUsed:                  UsedWhenRegistered,
				WantReplacementAttorneys:             "yes",
				HowReplacementAttorneysMakeDecisions: JointlyAndSeverally,
				HowShouldReplacementAttorneysStepIn:  OneCanNoLongerAct,
				ReplacementAttorneys: actor.Attorneys{
					{
						FirstNames: "Jane",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JaneSmith",
					},
					{
						FirstNames: "Jorge",
						LastName:   "Smith",
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JorgeSmith",
					},
				},
				You: actor.Person{
					FirstNames: "Jose",
					LastName:   "Smith",
					Address: place.Address{
						Line1:      "1 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
					Email:       TestEmail,
					DateOfBirth: date.New("2000", "1", "2"),
				},
				WhoFor: "me",
				Type:   LpaTypePropertyFinance,
				CertificateProvider: actor.CertificateProvider{
					FirstNames:              "Barbara",
					LastName:                "Smith",
					Email:                   TestEmail,
					Mobile:                  TestMobile,
					DateOfBirth:             date.New("1997", "1", "2"),
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
				},
				Attorneys: actor.Attorneys{
					{
						ID:          "JohnSmith",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:          "JoanSmith",
						FirstNames:  "Joan",
						LastName:    "Smith",
						Email:       TestEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						Address: place.Address{
							Line1:      "2 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				HowAttorneysMakeDecisions: JointlyAndSeverally,
				Tasks: Tasks{
					ConfirmYourIdentityAndSign: TaskCompleted,
					CheckYourLpa:               TaskCompleted,
					PeopleToNotify:             TaskCompleted,
					Restrictions:               TaskCompleted,
					WhenCanTheLpaBeUsed:        TaskCompleted,
					ChooseReplacementAttorneys: TaskCompleted,
					YourDetails:                TaskCompleted,
					CertificateProvider:        TaskCompleted,
					PayForLpa:                  TaskCompleted,
					ChooseAttorneys:            TaskCompleted,
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("as certificate provider", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&asCertificateProvider=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				CertificateProviderUserData: identity.UserData{
					FullName: "Barbara Smith",
					OK:       true,
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("provide certificate", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&provideCertificate=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &mockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := newMockLpaStore(t)
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				CertificateProviderUserData: identity.UserData{
					FullName: "Barbara Smith",
					OK:       true,
				},
				CertificateProviderProvidedDetails: actor.CertificateProvider{
					Mobile: TestMobile,
					Email:  TestEmail,
					Address: place.Address{
						Line1:      "5 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
				},
				Certificate: Certificate{
					AgreeToStatement: true,
					Agreed:           time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				},
			}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, &mockDataStore{}).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore)
	})

	t.Run("start certificate provider flow with identity", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&startCpFlowWithId=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &MockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &MockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
			}).
			Return(nil)

		dataStore := &mockDataStore{}
		dataStore.
			On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "MTIz", LpaID: "123", Identity: true}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, dataStore).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/certificate-provider-start?share-code=123", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore, dataStore)
	})

	t.Run("start certificate provider flow without identity", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&startCpFlowWithoutId=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionsStore := &MockSessionsStore{}
		sessionsStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpaStore := &MockLpaStore{}
		lpaStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)
		lpaStore.
			On("Put", ctx, &Lpa{
				ID: "123",
			}).
			Return(nil)

		dataStore := &mockDataStore{}
		dataStore.
			On("Put", ctx, "SHARECODE#123", "#METADATA#123", ShareCodeData{SessionID: "MTIz", LpaID: "123", Identity: false}).
			Return(nil)

		TestingStart(sessionsStore, lpaStore, MockRandom, dataStore).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/certificate-provider-start?share-code=123", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionsStore, lpaStore, dataStore)
	})
}
