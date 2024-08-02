package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestOrganisationStoreCreate(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Organisation{
			PK:        dynamo.OrganisationKey("a-uuid"),
			SK:        dynamo.OrganisationKey("a-uuid"),
			ID:        "a-uuid",
			CreatedAt: testNow,
			Name:      "A name",
		}).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	organisation, err := organisationStore.Create(ctx, &actor.Member{OrganisationID: "a-uuid"}, "A name")
	assert.Nil(t, err)
	assert.Equal(t, &actor.Organisation{
		PK:        dynamo.OrganisationKey("a-uuid"),
		SK:        dynamo.OrganisationKey("a-uuid"),
		ID:        "a-uuid",
		CreatedAt: testNow,
		Name:      "A name",
	}, organisation)
}

func TestOrganisationStoreCreateWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":   appcontext.ContextWithSession(context.Background(), &appcontext.Session{Email: "a@example.org"}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			organisation, err := organisationStore.Create(ctx, &actor.Member{OrganisationID: "a-uuid"}, "A name")
			assert.Error(t, err)
			assert.Nil(t, organisation)
		})
	}
}

func TestOrganisationStoreCreateWhenErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	organisation, err := organisationStore.Create(ctx, &actor.Member{OrganisationID: "a-uuid"}, "A name")
	assert.ErrorIs(t, err, expectedError)
	assert.Nil(t, organisation)
}

func TestOrganisationStoreGet(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	organisation := &actor.Organisation{Name: "A name"}

	member := actor.Member{PK: dynamo.OrganisationKey("a-uuid")}
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.MemberKey("session-id"), member, nil)
	dynamoClient.
		ExpectOne(ctx, dynamo.OrganisationKey("a-uuid"), dynamo.OrganisationKey("a-uuid"), organisation, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := organisationStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, organisation, result)
}

func TestOrganisationStoreGetWhenOrganisationDeleted(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	organisation := &actor.Organisation{Name: "A name", DeletedAt: testNow}

	member := actor.Member{PK: dynamo.OrganisationKey("a-uuid")}
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.MemberKey("session-id"), member, nil)
	dynamoClient.
		ExpectOne(ctx, dynamo.OrganisationKey("a-uuid"), dynamo.OrganisationKey("a-uuid"), organisation, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := organisationStore.Get(ctx)

	assert.Equal(t, dynamo.NotFoundError{}, err)
	assert.Nil(t, result)
}

func TestOrganisationStoreGetWithSessionErrors(t *testing.T) {
	testcases := map[string]context.Context{
		"missing":           context.Background(),
		"missing SessionID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
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

	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	member := actor.Member{PK: dynamo.OrganisationKey("a-uuid")}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneBySK(ctx, dynamo.MemberKey("session-id"), member, tc.oneBySKError)

			if tc.oneError != nil {
				dynamoClient.
					ExpectOne(ctx, dynamo.OrganisationKey("a-uuid"), dynamo.OrganisationKey("a-uuid"), nil, tc.oneError)
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
		Put(ctx, &actor.Organisation{PK: dynamo.OrganisationKey("123"), SK: dynamo.OrganisationKey("456"), Name: "Hey", UpdatedAt: testNow}).
		Return(expectedError)

	store := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Organisation{PK: dynamo.OrganisationKey("123"), SK: dynamo.OrganisationKey("456"), Name: "Hey"})
	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreCreateLPA(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})
	expectedDonor := &donordata.Provided{
		PK:        dynamo.LpaKey("a-uuid"),
		SK:        dynamo.LpaOwnerKey(dynamo.OrganisationKey("an-id")),
		LpaID:     "a-uuid",
		CreatedAt: testNow,
		Version:   1,
		Donor: donordata.Donor{
			UID: testUID,
		},
	}
	expectedDonor.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, expectedDonor).
		Return(nil)

	organisationStore := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
		uuidString:   func() string { return "a-uuid" },
		newUID:       testUIDFn,
	}

	donor, err := organisationStore.CreateLPA(ctx)

	assert.Nil(t, err)
	assert.Equal(t, expectedDonor, donor)
}

func TestOrganisationStoreCreateLPAWithSessionMissing(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":         context.Background(),
		"missing organisation ID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
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
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	organisationStore := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
		uuidString:   func() string { return "a-uuid" },
		newUID:       testUIDFn,
	}

	_, err := organisationStore.CreateLPA(ctx)

	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreSoftDelete(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id", SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.Organisation{DeletedAt: testNow}).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	err := organisationStore.SoftDelete(ctx, &actor.Organisation{})
	assert.Nil(t, err)
}

func TestOrganisationStoreSoftDeleteWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id", SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	err := organisationStore.SoftDelete(ctx, &actor.Organisation{})
	assert.Error(t, err)
}
