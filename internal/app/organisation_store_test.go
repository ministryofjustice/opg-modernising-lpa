package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestOrganisationStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Organisation{
			PK:        "ORGANISATION#a-uuid",
			SK:        "ORGANISATION#a-uuid",
			ID:        "a-uuid",
			CreatedAt: testNow,
			Name:      "A name",
		}).
		Return(nil).
		Once()

	dynamoClient.EXPECT().
		Create(ctx, &actor.Member{
			PK:        "ORGANISATION#a-uuid",
			SK:        "MEMBER#an-id",
			ID:        "a-uuid",
			CreatedAt: testNow,
			Email:     "a@example.org",
		}).
		Return(nil).
		Once()

	dynamoClient.EXPECT().
		Create(ctx, &organisationLink{
			PK:       "ORGANISATION#a-uuid",
			SK:       "MEMBERID#a-uuid",
			MemberSK: "MEMBER#an-id",
		}).
		Return(nil).
		Once()

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	organisation, err := organisationStore.Create(ctx, "A name")
	assert.Nil(t, err)
	assert.Equal(t, &actor.Organisation{
		PK:        "ORGANISATION#a-uuid",
		SK:        "ORGANISATION#a-uuid",
		ID:        "a-uuid",
		CreatedAt: testNow,
		Name:      "A name",
	}, organisation)
}

func TestOrganisationStoreCreateWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":   page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "a@example.org"}),
		"no email":        page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			organisation, err := organisationStore.Create(ctx, "A name")
			assert.Error(t, err)
			assert.Nil(t, organisation)
		})
	}
}

func TestOrganisationStoreCreateWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", Email: "a@example.org"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"organisation": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"member": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDynamoClient(t)
			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			organisation, err := organisationStore.Create(ctx, "A name")
			assert.ErrorIs(t, err, expectedError)
			assert.Nil(t, organisation)
		})
	}
}

func TestOrganisationStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	organisation := &actor.Organisation{Name: "A name"}

	member := actor.Member{PK: "ORGANISATION#a-uuid"}
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, "MEMBER#session-id", member, nil)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "ORGANISATION#a-uuid", organisation, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := organisationStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, organisation, result)
}

func TestOrganisationStoreGetWithSessionErrors(t *testing.T) {
	testcases := map[string]context.Context{
		"missing":           context.Background(),
		"missing SessionID": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			_, err := organisationStore.Get(ctx)
			assert.Error(t, err)
		})
	}
}

func TestOrganisationStoreGetWhenErrors(t *testing.T) {
	testcases := map[string]struct {
		oneBySKError error
		oneError     error
	}{
		"OneBySK error": {
			oneBySKError: expectedError,
		},
		"One error": {
			oneError: expectedError,
		},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	member := actor.Member{PK: "ORGANISATION#a-uuid"}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneBySK(ctx, "MEMBER#session-id", member, tc.oneBySKError)

			if tc.oneError != nil {
				dynamoClient.
					ExpectOne(ctx, "ORGANISATION#a-uuid", "ORGANISATION#a-uuid", nil, tc.oneError)
			}

			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := organisationStore.Get(ctx)
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestOrganisationStorePut(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.Organisation{PK: "ORGANISATION#123", SK: "ORGANISATION#456", Name: "Hey", UpdatedAt: testNow}).
		Return(expectedError)

	store := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Organisation{PK: "ORGANISATION#123", SK: "ORGANISATION#456", Name: "Hey"})
	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreCreateLPA(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})
	expectedDonor := &actor.DonorProvidedDetails{
		PK:        "LPA#a-uuid",
		SK:        "ORGANISATION#an-id",
		LpaID:     "a-uuid",
		CreatedAt: testNow,
		Version:   1,
	}
	expectedDonor.Hash, _ = expectedDonor.GenerateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, expectedDonor).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	donor, err := organisationStore.CreateLPA(ctx)

	assert.Nil(t, err)
	assert.Equal(t, expectedDonor, donor)
}

func TestOrganisationStoreCreateLPAWithSessionMissing(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":         context.Background(),
		"missing organisation ID": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
	}

	for name, ctx := range testCases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := organisationStore.CreateLPA(ctx)
			assert.Error(t, err)
		})
	}
}

func TestOrganisationStoreCreateLPAWhenDynamoError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.CreateLPA(ctx)

	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreAllLPAs(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})
	expectedDonorA := actor.DonorProvidedDetails{
		PK:     "LPA#a-uuid",
		SK:     "ORGANISATION#an-id",
		LpaUID: "a-uid",
		Donor: actor.Donor{
			FirstNames: "a",
			LastName:   "a",
		},
	}
	expectedDonorB := actor.DonorProvidedDetails{
		PK:     "LPA#b-uuid",
		SK:     "ORGANISATION#an-id",
		LpaUID: "b-uid",
		Donor: actor.Donor{
			FirstNames: "a",
			LastName:   "b",
		},
	}
	expectedDonorC := actor.DonorProvidedDetails{
		PK:     "LPA#c-uuid",
		SK:     "ORGANISATION#an-id",
		LpaUID: "c-uid",
		Donor: actor.Donor{
			FirstNames: "c",
			LastName:   "a",
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, "ORGANISATION#an-id",
		[]actor.DonorProvidedDetails{
			expectedDonorB,
			expectedDonorC,
			expectedDonorA,
			{PK: "ORGANISATION#an-id", SK: "ORGANISATION#an-id"},
			{
				PK:    "LPA#d-uuid",
				SK:    "ORGANISATION#an-id",
				LpaID: "d-uuid",
				Donor: actor.Donor{
					FirstNames: "d",
					LastName:   "d",
				},
			},
		}, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	donors, err := organisationStore.AllLPAs(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []actor.DonorProvidedDetails{expectedDonorA, expectedDonorB, expectedDonorC}, donors)
}

func TestOrganisationStoreAllLPAsWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":   page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			donors, err := organisationStore.AllLPAs(ctx)
			assert.Error(t, err)
			assert.Nil(t, donors)
		})
	}
}

func TestOrganisationStoreAllLPAsWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, "ORGANISATION#an-id",
		nil, expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.AllLPAs(ctx)
	assert.ErrorIs(t, err, expectedError)
}
