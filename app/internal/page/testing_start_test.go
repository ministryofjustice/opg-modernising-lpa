package page

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestingStart(t *testing.T) {
	t.Run("with type", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?withType=hw&redirect=/start", nil)
		now := time.Now()

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})).
			Return(&Lpa{ID: "123"}, nil)
		donorStore.
			On("Put", ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"}), &Lpa{ID: "123", Type: "hw"}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, func() time.Time { return now }).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/start", resp.Header.Get("Location"))
	})

	t.Run("payment not complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{ID: "123"}).
			Return(nil)

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("payment complete", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&paymentComplete=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx,
				mock.MatchedBy(func(lpa *Lpa) bool {
					return assert.Equal(t, actor.TaskCompleted, lpa.Tasks.PayForLpa) &&
						assert.Equal(t, PaymentDetails{PaymentReference: "123", PaymentId: "123"}, lpa.PaymentDetails)
				}),
			).
			Return(nil)

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("with attorney", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withAttorney=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				Attorneys: actor.Attorneys{
					{
						ID:          "JohnSmith",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       testEmail,
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
					ChooseAttorneys: actor.TaskCompleted,
				},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
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
				Email:       testEmail,
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
				Email:       testEmail,
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{},
			},
		}

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                                  "123",
				Type:                                LpaTypePropertyFinance,
				WhenCanTheLpaBeUsed:                 UsedWhenRegistered,
				Attorneys:                           attorneys,
				ReplacementAttorneys:                attorneys,
				AttorneyDecisions:                   actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
				WantReplacementAttorneys:            "yes",
				ReplacementAttorneyDecisions:        actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
				HowShouldReplacementAttorneysStepIn: OneCanNoLongerAct,
				Tasks: Tasks{
					ChooseAttorneys:            actor.TaskInProgress,
					ChooseReplacementAttorneys: actor.TaskInProgress,
				},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
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
				Email:       testEmail,
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
				Email:       testEmail,
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

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                "123",
				Attorneys:         attorneys,
				AttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
				Tasks: Tasks{
					ChooseAttorneys: actor.TaskCompleted,
				},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
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

				sessionStore := newMockSessionStore(t)
				sessionStore.
					On("Save", r, w, mock.Anything).
					Return(nil)

				donorStore := newMockDonorStore(t)
				donorStore.
					On("Create", ctx).
					Return(&Lpa{ID: "123"}, nil)

				ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

				donorStore.
					On("Put", ctx, &Lpa{
						ID: "123",
						AttorneyDecisions: actor.AttorneyDecisions{
							How:     tc.DecisionsType,
							Details: tc.DecisionsDetails,
						},
					}).
					Return(nil)

				TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))

			})
		}
	})

	t.Run("how replacement attorneys act", func(t *testing.T) {
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
				r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/?redirect=/somewhere&howReplacementAttorneysAct=%s", tc.DecisionsType), nil)
				ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

				sessionStore := newMockSessionStore(t)
				sessionStore.
					On("Save", r, w, mock.Anything).
					Return(nil)

				donorStore := newMockDonorStore(t)
				donorStore.
					On("Create", ctx).
					Return(&Lpa{ID: "123"}, nil)

				ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

				donorStore.
					On("Put", ctx, &Lpa{
						ID: "123",
						ReplacementAttorneyDecisions: actor.AttorneyDecisions{
							How:                        tc.DecisionsType,
							Details:                    tc.DecisionsDetails,
							HappyIfOneCannotActNoneCan: "yes",
						},
					}).
					Return(nil)

				TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
				resp := w.Result()

				assert.Equal(t, http.StatusFound, resp.StatusCode)
				assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))

			})
		}
	})

	t.Run("with Certificate Provider", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withCPDetails=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				CertificateProvider: actor.CertificateProvider{
					FirstNames:              "Jessie",
					LastName:                "Jones",
					Email:                   testEmail,
					Mobile:                  testMobile,
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
					CarryOutBy:              "paper",
					Address: place.Address{
						Line1:      "5 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
				},
				Tasks: Tasks{CertificateProvider: actor.TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("with donor details", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withDonorDetails=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				Donor: actor.Donor{
					FirstNames: "Jamie",
					LastName:   "Smith",
					Address: place.Address{
						Line1:      "1 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
					Email:       testEmail,
					DateOfBirth: date.New("2000", "1", "2"),
				},
				WhoFor: "me",
				Type:   LpaTypePropertyFinance,
				Tasks:  Tasks{YourDetails: actor.TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("with replacement attorneys", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withReplacementAttorneys=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                       "123",
				WantReplacementAttorneys: "yes",
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{
					How: actor.JointlyAndSeverally,
				},
				HowShouldReplacementAttorneysStepIn: OneCanNoLongerAct,
				Tasks:                               Tasks{ChooseReplacementAttorneys: actor.TaskCompleted},
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
						Email:       testEmail,
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
						Email:       testEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JorgeSmith",
					},
				},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("when can be used completed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&whenCanBeUsedComplete=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                  "123",
				WhenCanTheLpaBeUsed: UsedWhenRegistered,
				Tasks:               Tasks{WhenCanTheLpaBeUsed: actor.TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("with restrictions", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withRestrictions=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:           "123",
				Restrictions: "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently",
				Tasks:        Tasks{Restrictions: actor.TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("with people to notify", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withPeopleToNotify=5", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                      "123",
				DoYouWantToNotifyPeople: "yes",
				Tasks:                   Tasks{PeopleToNotify: actor.TaskCompleted},
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      testEmail,
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
						Email:      testEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JulianSmith",
						FirstNames: "Julian",
						LastName:   "Smith",
						Email:      testEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JaydenSmith",
						FirstNames: "Jayden",
						LastName:   "Smith",
						Email:      testEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
					{
						ID:         "JuniperSmith",
						FirstNames: "Juniper",
						LastName:   "Smith",
						Email:      testEmail,
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

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("with incomplete people to notify", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&withIncompletePeopleToNotify=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                      "123",
				DoYouWantToNotifyPeople: "yes",
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      testEmail,
						Address:    place.Address{},
					},
				},
				Tasks: Tasks{PeopleToNotify: actor.TaskInProgress},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("lpa checked", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&lpaChecked=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:           "123",
				Checked:      true,
				HappyToShare: true,
				Tasks:        Tasks{CheckYourLpa: actor.TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("id confirmed and signed", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&idConfirmedAndSigned=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				DonorIdentityUserData: identity.UserData{
					OK:          true,
					Provider:    identity.OneLogin,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FirstNames:  "Jamie",
					LastName:    "Smith",
				},
				WantToApplyForLpa:      true,
				WantToSignLpa:          true,
				CPWitnessCodeValidated: true,
				Submitted:              time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				Tasks:                  Tasks{ConfirmYourIdentityAndSign: actor.TaskCompleted},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("complete LPA", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&completeLpa=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID: "123",
				DonorIdentityUserData: identity.UserData{
					OK:          true,
					Provider:    identity.OneLogin,
					RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
					FirstNames:  "Jamie",
					LastName:    "Smith",
				},
				WantToApplyForLpa:      true,
				WantToSignLpa:          true,
				CPWitnessCodeValidated: true,
				PaymentDetails: PaymentDetails{
					PaymentReference: "123",
					PaymentId:        "123",
				},
				Submitted:               time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
				Checked:                 true,
				HappyToShare:            true,
				DoYouWantToNotifyPeople: "yes",
				PeopleToNotify: actor.PeopleToNotify{
					{
						ID:         "JoannaSmith",
						FirstNames: "Joanna",
						LastName:   "Smith",
						Email:      testEmail,
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
						Email:      testEmail,
						Address: place.Address{
							Line1:      "4 RICHMOND PLACE",
							Line2:      "KINGS HEATH",
							Line3:      "WEST MIDLANDS",
							TownOrCity: "BIRMINGHAM",
							Postcode:   "B14 7ED",
						},
					},
				},
				Restrictions:             "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently",
				WhenCanTheLpaBeUsed:      UsedWhenRegistered,
				WantReplacementAttorneys: "yes",
				ReplacementAttorneyDecisions: actor.AttorneyDecisions{
					How: actor.JointlyAndSeverally,
				},
				HowShouldReplacementAttorneysStepIn: OneCanNoLongerAct,
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
						Email:       testEmail,
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
						Email:       testEmail,
						DateOfBirth: date.New("2000", "1", "2"),
						ID:          "JorgeSmith",
					},
				},
				Donor: actor.Donor{
					FirstNames: "Jamie",
					LastName:   "Smith",
					Address: place.Address{
						Line1:      "1 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
					Email:       testEmail,
					DateOfBirth: date.New("2000", "1", "2"),
				},
				WhoFor: "me",
				Type:   LpaTypePropertyFinance,
				CertificateProvider: actor.CertificateProvider{
					FirstNames:              "Jessie",
					LastName:                "Jones",
					Email:                   testEmail,
					Mobile:                  testMobile,
					Relationship:            "friend",
					RelationshipDescription: "",
					RelationshipLength:      "gte-2-years",
					CarryOutBy:              "paper",
					Address: place.Address{
						Line1:      "5 RICHMOND PLACE",
						Line2:      "KINGS HEATH",
						Line3:      "WEST MIDLANDS",
						TownOrCity: "BIRMINGHAM",
						Postcode:   "B14 7ED",
					},
				},
				Attorneys: actor.Attorneys{
					{
						ID:          "JohnSmith",
						FirstNames:  "John",
						LastName:    "Smith",
						Email:       testEmail,
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
						Email:       testEmail,
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
				AttorneyDecisions: actor.AttorneyDecisions{
					How: actor.JointlyAndSeverally,
				},
				Tasks: Tasks{
					ConfirmYourIdentityAndSign: actor.TaskCompleted,
					CheckYourLpa:               actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					WhenCanTheLpaBeUsed:        actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					YourDetails:                actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PayForLpa:                  actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
				},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("as certificate provider", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&asCertificateProvider=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID: "123",
			}).
			Return(nil)

		certificateProviderStore := newMockCertificateProviderStore(t)
		certificateProviderStore.
			On("Create", ctx, "MTIz").
			Return(&actor.CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{
					OK:         true,
					Provider:   identity.OneLogin,
					FirstNames: "Jessie",
					LastName:   "Jones",
				},
			}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{
			SessionID: base64.StdEncoding.EncodeToString([]byte("123")),
			LpaID:     "123",
		})

		certificateProviderStore.
			On("Put", ctx, &actor.CertificateProviderProvidedDetails{
				IdentityUserData: identity.UserData{
					OK:         true,
					Provider:   identity.OneLogin,
					FirstNames: "Jessie",
					LastName:   "Jones",
				},
			}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, certificateProviderStore, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/lpa/123/somewhere", resp.Header.Get("Location"))
	})

	t.Run("start certificate provider flow - donor has paid", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&startCpFlowDonorHasPaid=1&useTestShareCode=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpa := &Lpa{
			ID:                  "123",
			CertificateProvider: actor.CertificateProvider{Email: testEmail},
		}
		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(lpa, nil)

		lpa.PaymentDetails = PaymentDetails{
			PaymentReference: "123",
			PaymentId:        "123",
		}

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, lpa).
			Return(nil)

		localizer := newMockLocalizer(t)

		shareCodeSender := newMockShareCodeSender(t)
		shareCodeSender.
			On("UseTestCode").
			Return(nil)
		shareCodeSender.
			On("SendCertificateProvider", ctx, notify.CertificateProviderInviteEmail, AppData{SessionID: "MTIz", LpaID: "123", Localizer: localizer}, false, lpa).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, shareCodeSender, localizer, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/certificate-provider-start", resp.Header.Get("Location"))
	})

	t.Run("start certificate provider flow - donor has not paid", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&startCpFlowDonorHasNotPaid=1&useTestShareCode=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpa := &Lpa{
			ID:                  "123",
			CertificateProvider: actor.CertificateProvider{Email: testEmail},
		}
		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(lpa, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, lpa).
			Return(nil)

		localizer := newMockLocalizer(t)

		shareCodeSender := newMockShareCodeSender(t)
		shareCodeSender.
			On("UseTestCode").
			Return(nil)
		shareCodeSender.
			On("SendCertificateProvider", ctx, notify.CertificateProviderInviteEmail, AppData{SessionID: "MTIz", LpaID: "123", Localizer: localizer}, false, lpa).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, shareCodeSender, localizer, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/certificate-provider-start", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionStore, donorStore, shareCodeSender)
	})

	t.Run("start certificate provider flow with email", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/somewhere&startCpFlowDonorHasNotPaid=1&useTestShareCode=1&withEmail=a@example.org", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		lpa := &Lpa{
			ID:                  "123",
			CertificateProvider: actor.CertificateProvider{Email: testEmail},
		}

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(lpa, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, lpa).
			Return(nil)

		localizer := newMockLocalizer(t)

		shareCodeSender := newMockShareCodeSender(t)
		shareCodeSender.
			On("UseTestCode").
			Return(nil)
		shareCodeSender.
			On("SendCertificateProvider", ctx, notify.CertificateProviderInviteEmail, AppData{SessionID: "MTIz", LpaID: "123", Localizer: localizer}, false, lpa).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, shareCodeSender, localizer, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/certificate-provider-start", resp.Header.Get("Location"))
		mock.AssertExpectationsForObjects(t, sessionStore, donorStore, shareCodeSender)
	})

	t.Run("as attorney", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/attorney-start&asAttorney=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123", Attorneys: actor.Attorneys{{ID: "456"}}}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:        "123",
				Attorneys: actor.Attorneys{{ID: "456"}},
			}).
			Return(nil)

		attorneyStore := newMockAttorneyStore(t)
		attorneyStore.
			On("Create", ctx, "MTIz", "456", false).
			Return(nil, nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, attorneyStore, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})

	t.Run("as replacement attorney", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?redirect=/attorney-start&asReplacementAttorney=1", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123", ReplacementAttorneys: actor.Attorneys{{ID: "456"}}}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, &Lpa{
				ID:                   "123",
				ReplacementAttorneys: actor.Attorneys{{ID: "456"}},
			}).
			Return(nil)

		attorneyStore := newMockAttorneyStore(t)
		attorneyStore.
			On("Create", ctx, "MTIz", "456", true).
			Return(nil, nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, attorneyStore, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})

	t.Run("send attorney share", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?sendAttorneyShare=1&redirect=/attorney-start", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		lpa := &Lpa{ID: "123"}

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		localizer := newMockLocalizer(t)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(lpa, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})
		donorStore.
			On("Put", ctx, mock.Anything).
			Return(nil)

		shareCodeSender := newMockShareCodeSender(t)
		shareCodeSender.
			On("SendAttorneys", ctx, AppData{SessionID: "MTIz", LpaID: "123", Localizer: localizer}, lpa).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, shareCodeSender, localizer, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})

	t.Run("send attorney share with email", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?sendAttorneyShare=1&withEmail=a@b.c&redirect=/attorney-start", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		lpa := &Lpa{ID: "123", Attorneys: actor.Attorneys{makeAttorney(attorneyNames[0])}}

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		localizer := newMockLocalizer(t)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(lpa, nil)

		lpa.Attorneys[0].Email = "a@b.c"
		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, mock.Anything).
			Return(nil)

		shareCodeSender := newMockShareCodeSender(t)
		shareCodeSender.
			On("SendAttorneys", ctx, AppData{SessionID: "MTIz", LpaID: "123", Localizer: localizer}, lpa).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, shareCodeSender, localizer, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})

	t.Run("send replacement attorney share", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?sendAttorneyShare=1&forReplacementAttorney=1&redirect=/attorney-start", nil)
		ctx := ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})

		lpa := &Lpa{ID: "123"}

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		localizer := newMockLocalizer(t)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ctx).
			Return(&Lpa{ID: "123"}, nil)

		ctx = ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"})

		donorStore.
			On("Put", ctx, lpa).
			Return(nil)

		shareCodeSender := newMockShareCodeSender(t)
		shareCodeSender.
			On("SendAttorneys", ctx, AppData{SessionID: "MTIz", LpaID: "123", Localizer: localizer}, lpa).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, shareCodeSender, localizer, nil, nil, nil, nil).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})

	t.Run("signed by donor", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?signedByDonor=1&redirect=/attorney-start", nil)
		now := time.Now()

		sessionStore := newMockSessionStore(t)
		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})).
			Return(&Lpa{ID: "123"}, nil)
		donorStore.
			On("Put", ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"}), &Lpa{ID: "123", Submitted: now}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, func() time.Time { return now }).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})

	t.Run("with share code sessionr", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "/?withShareCodeSession=1&redirect=/attorney-start", nil)
		now := time.Now()

		sessionStore := newMockSessionStore(t)

		session := sessions.NewSession(sessionStore, "shareCode")

		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400,
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
			Secure:   true,
		}
		session.Values = map[any]any{"share-code": &sesh.ShareCodeSession{LpaID: "123", Identity: false}}

		sessionStore.
			On("Save", r, w, mock.Anything).
			Return(nil)

		sessionStore.
			On("Save", r, w, session).
			Return(nil)

		donorStore := newMockDonorStore(t)
		donorStore.
			On("Create", ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz"})).
			Return(&Lpa{ID: "123"}, nil)
		donorStore.
			On("Put", ContextWithSessionData(r.Context(), &SessionData{SessionID: "MTIz", LpaID: "123"}), &Lpa{ID: "123"}).
			Return(nil)

		TestingStart(sessionStore, donorStore, MockRandom, nil, nil, nil, nil, nil, func() time.Time { return now }).ServeHTTP(w, r)
		resp := w.Result()

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/attorney-start", resp.Header.Get("Location"))
	})
}
