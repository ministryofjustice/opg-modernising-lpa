package donordata

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestDonorStoreGetAny(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey(""), &Provided{LpaID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWhenOrganisation(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", OrganisationID: "x"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.OrganisationKey(""), &Provided{LpaID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWhenReference(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey(""),
		lpaReference{
			PK:           dynamo.LpaKey("an-id"),
			SK:           dynamo.DonorKey("donor"),
			ReferencedSK: dynamo.OrganisationKey("org"),
		}, nil)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.OrganisationKey("org"),
		&Provided{LpaID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &Provided{SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreGetAnyWhenDataStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey(""), &Provided{LpaID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGetAnyWhenReferenceErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey(""),
		lpaReference{
			PK:           dynamo.LpaKey("an-id"),
			SK:           dynamo.DonorKey("donor"),
			ReferencedSK: dynamo.OrganisationKey("org"),
		}, nil)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.OrganisationKey("org"),
		nil, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGet(t *testing.T) {
	testCases := map[string]struct {
		sessionData *appcontext.Session
		expectedSK  dynamo.SK
	}{
		"donor": {
			sessionData: &appcontext.Session{LpaID: "an-id", SessionID: "456"},
			expectedSK:  dynamo.DonorKey("456"),
		},
		"organisation": {
			sessionData: &appcontext.Session{LpaID: "an-id", SessionID: "456", OrganisationID: "789"},
			expectedSK:  dynamo.OrganisationKey("789"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), tc.sessionData)

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), tc.expectedSK, &Provided{LpaID: "an-id"}, nil)

			donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

			lpa, err := donorStore.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, &Provided{LpaID: "an-id"}, lpa)
		})
	}
}

func TestDonorStoreGetWhenReferenced(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey("456"), lpaReference{ReferencedSK: dynamo.OrganisationKey("789")}, nil)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.OrganisationKey("789"), &Provided{LpaID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, &Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreGetWhenDataStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey("456"), lpaReference{ReferencedSK: "ref"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreLatest(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, dynamo.DonorKey("456"), &Provided{LpaID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Latest(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreLatestWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreLatestWhenDataStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, dynamo.DonorKey("456"), &Provided{LpaID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGetByKeys(t *testing.T) {
	keys := []dynamo.Keys{
		{PK: dynamo.LpaKey("1"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a"))},
		{PK: dynamo.LpaKey("2"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("b"))},
		{PK: dynamo.LpaKey("3"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("c"))},
	}
	donors := []Provided{
		{PK: dynamo.LpaKey("1"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a")), LpaID: "1"},
		{PK: dynamo.LpaKey("2"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("b")), LpaID: "2"},
		{PK: dynamo.LpaKey("3"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("c")), LpaID: "3"},
	}
	av0, _ := attributevalue.MarshalMap(donors[2])
	av1, _ := attributevalue.MarshalMap(donors[1])
	av2, _ := attributevalue.MarshalMap(donors[0])

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByKeys(ctx, keys,
		[]map[string]types.AttributeValue{av0, av1, av2}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Equal(t, donors, result)
}

func TestDonorStoreGetByKeysWhenNoKeys(t *testing.T) {
	keys := []dynamo.Keys{}

	donorStore := &donorStore{}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestDonorStoreGetByKeysWhenDynamoErrors(t *testing.T) {
	keys := []dynamo.Keys{{}}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByKeys(ctx, keys,
		nil, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	_, err := donorStore.GetByKeys(ctx, keys)
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePut(t *testing.T) {
	saved := &Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: Donor{FirstNames: "x", LastName: "y"}}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, &Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: Donor{FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDSet(t *testing.T) {
	saved := &Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", UpdatedAt: testNow, Donor: Donor{FirstNames: "x", LastName: "y"}}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, search.Lpa{PK: dynamo.LpaKey("5").PK(), SK: dynamo.DonorKey("an-id").SK(), Donor: search.LpaDonor{FirstNames: "x", LastName: "y"}}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", Donor: Donor{FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDSetIndexErrors(t *testing.T) {
	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.EXPECT().
		WarnContext(ctx, "donorStore index failed", slog.Any("err", expectedError))

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, searchClient: searchClient, logger: logger, now: testNowFn}

	err := donorStore.Put(ctx, &Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", Donor: Donor{FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenNoChange(t *testing.T) {
	donorStore := &donorStore{}

	donor := &Provided{LpaID: "an-id"}
	donor.Hash, _ = hashstructure.Hash(donor, hashstructure.FormatV2, nil)

	err := donorStore.Put(ctx, donor)
	assert.Nil(t, err)
}

func TestDonorStorePutWhenCheckChangeAndCheckCompleted(t *testing.T) {
	saved := &Provided{PK: dynamo.LpaKey("5"), Hash: 5, CheckedHash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: Donor{FirstNames: "a", LastName: "b"}, Tasks: Tasks{CheckYourLpa: task.StateInProgress}}
	_ = saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, &Provided{PK: dynamo.LpaKey("5"), Hash: 5, CheckedHash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: Donor{FirstNames: "a", LastName: "b"}, Tasks: Tasks{CheckYourLpa: task.StateCompleted}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().Put(ctx, mock.Anything).Return(expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: time.Now}

	err := donorStore.Put(ctx, &Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5"})
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenApplicationUpdatedWhenError(t *testing.T) {
	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationUpdated(ctx, mock.Anything).
		Return(expectedError)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{eventClient: eventClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &Provided{
		PK:     dynamo.LpaKey("5"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		LpaID:  "5",
		LpaUID: "M-1111",
		Donor: Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Postcode: "F1 1FF",
			},
		},
		Type: lpadata.LpaTypePersonalWelfare,
	})

	assert.Equal(t, expectedError, err)
}

func TestDonorStoreCreate(t *testing.T) {
	testCases := map[string]Provided{
		"with previous details": {
			Donor: Donor{
				UID:         actoruid.New(),
				FirstNames:  "a",
				LastName:    "b",
				OtherNames:  "c",
				DateOfBirth: date.New("2000", "01", "02"),
				Address:     place.Address{Line1: "d"},
			},
		},
		"no previous details": {},
	}

	for name, previousDetails := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})
			donor := &Provided{
				PK:        dynamo.LpaKey("10100000"),
				SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
				LpaID:     "10100000",
				CreatedAt: testNow,
				Version:   1,
				Donor: Donor{
					UID:         testUID,
					FirstNames:  previousDetails.Donor.FirstNames,
					LastName:    previousDetails.Donor.LastName,
					OtherNames:  previousDetails.Donor.OtherNames,
					DateOfBirth: previousDetails.Donor.DateOfBirth,
					Address:     previousDetails.Donor.Address,
					Channel:     lpadata.ChannelOnline,
				},
			}
			donor.UpdateHash()

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), previousDetails, nil)
			dynamoClient.EXPECT().
				Create(ctx, donor).
				Return(nil)
			dynamoClient.EXPECT().
				Create(ctx, actor.LpaLink{PK: dynamo.LpaKey("10100000"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor, UpdatedAt: testNow}).
				Return(nil)

			donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }, now: testNowFn, newUID: testUIDFn}

			result, err := donorStore.Create(ctx)
			assert.Nil(t, err)
			assert.Equal(t, donor, result)
		})
	}
}

func TestDonorStoreCreateWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }, now: func() time.Time { return time.Now() }}

	_, err := donorStore.Create(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreCreateWhenError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"latest": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), Provided{}, expectedError)

			return dynamoClient
		},
		"donor record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), Provided{}, nil)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"link record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), Provided{}, nil)
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

	for name, makeMockDataStore := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDataStore(t)

			donorStore := &donorStore{
				dynamoClient: dynamoClient,
				uuidString:   func() string { return "10100000" },
				now:          testNowFn,
				newUID:       testUIDFn,
			}

			_, err := donorStore.Create(ctx)
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestDonorStoreLink(t *testing.T) {
	testcases := map[string]struct {
		oneByPartialSKError error
		link                actor.LpaLink
	}{
		"no link": {
			oneByPartialSKError: dynamo.NotFoundError{},
		},
		"not a donor link": {
			link: actor.LpaLink{
				PK:        dynamo.LpaKey(""),
				SK:        dynamo.SubKey("a-sub"),
				ActorType: actor.TypeCertificateProvider,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
			shareCode := sharecode.Data{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
			}

			expectedTransaction := &dynamo.Transaction{
				Creates: []any{
					lpaReference{
						PK:           dynamo.LpaKey("lpa-id"),
						SK:           dynamo.DonorKey("session-id"),
						ReferencedSK: dynamo.OrganisationKey("org-id"),
					},
					actor.LpaLink{
						PK:        dynamo.LpaKey("lpa-id"),
						SK:        dynamo.SubKey("session-id"),
						DonorKey:  dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
						ActorType: actor.TypeDonor,
						UpdatedAt: testNowFn(),
					},
				},
				Puts: []any{
					sharecode.Data{
						LpaKey:      dynamo.LpaKey("lpa-id"),
						LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
						LpaLinkedTo: "a@example.com",
						LpaLinkedAt: testNowFn(),
					},
				},
			}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), tc.link, tc.oneByPartialSKError)
			dynamoClient.EXPECT().
				WriteTransaction(ctx, expectedTransaction).
				Return(nil)

			donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

			err := donorStore.Link(ctx, shareCode, "a@example.com")
			assert.Nil(t, err)
		})
	}
}

func TestDonorStoreLinkWithDonor(t *testing.T) {
	donorStore := &donorStore{}

	err := donorStore.Link(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor"))}, "a@example.com")
	assert.Error(t, err)
}

func TestDonorStoreLinkWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{}

	err := donorStore.Link(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))}, "a@example.com")
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreLinkWithSessionIDMissing(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})
	donorStore := &donorStore{}

	err := donorStore.Link(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))}, "a@example.com")
	assert.Error(t, err)
}

func TestDonorStoreLinkWhenDonorLinkAlreadyExists(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), actor.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("a-sub"), ActorType: actor.TypeDonor}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Link(
		ctx,
		sharecode.Data{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))},
		"a@example.com",
	)

	assert.Equal(t, errors.New("a donor link already exists for lpa-id"), err)
}

func TestDonorStoreLinkWhenError(t *testing.T) {
	testcases := map[string]func(*mockDynamoClient){
		"OneByPartialSK errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.
				ExpectOneByPartialSK(mock.Anything, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), actor.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("a-sub"), ActorType: actor.TypeAttorney}, expectedError)
		},
		"WriteTransaction errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.
				ExpectOneByPartialSK(mock.Anything, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), actor.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("a-sub"), ActorType: actor.TypeAttorney}, nil)
			dynamoClient.EXPECT().
				WriteTransaction(mock.Anything, mock.Anything).
				Return(expectedError)
		},
	}

	for name, setupDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})
			shareCode := sharecode.Data{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
			}

			dynamoClient := newMockDynamoClient(t)
			setupDynamoClient(dynamoClient)

			donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

			err := donorStore.Link(ctx, shareCode, "a@example.com")
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestDonorStoreDelete(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "123"})

	keys := []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("sk1")},
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("sk2")},
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, dynamo.LpaKey("123")).
		Return(keys, nil)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, keys).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.Nil(t, err)
}

func TestDonorStoreDeleteWhenOtherDonor(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "123"})

	keys := []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("sk1")},
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("sk2")},
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("another-id")},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, dynamo.LpaKey("123")).
		Return(keys, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.NotNil(t, err)
}

func TestDonorStoreDeleteWhenAllKeysByPKErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, dynamo.LpaKey("123")).
		Return(nil, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreDeleteWhenDeleteKeysErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, dynamo.LpaKey("123")).
		Return([]dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")}}, nil)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreDeleteWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"missing":      context.Background(),
		"no LpaID":     appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"}),
		"no SessionID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"}),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			donorStore := &donorStore{}

			err := donorStore.Delete(ctx)
			assert.NotNil(t, err)
		})
	}
}

func TestDonorStoreDeleteDonorAccess(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "org-id"})

	link := actor.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("donor-sub")}
	shareCodeData := sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), link, nil)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: link.PK, SK: link.SK},
				{PK: shareCodeData.LpaKey, SK: dynamo.DonorKey(link.UserSub())},
				{PK: shareCodeData.PK, SK: shareCodeData.SK},
			},
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.DeleteDonorAccess(ctx, shareCodeData)
	assert.Nil(t, err)
}

func TestDonorStoreDeleteDonorAccessWhenDonor(t *testing.T) {
	donorStore := &donorStore{}

	err := donorStore.DeleteDonorAccess(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor"))})
	assert.Error(t, err)
}

func TestDonorStoreDeleteDonorAccessWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"missing":           context.Background(),
		"no organisationID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"}),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			donorStore := &donorStore{}

			err := donorStore.DeleteDonorAccess(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))})
			assert.Error(t, err)
		})
	}
}

func TestDonorStoreDeleteDonorAccessWhenDeleterInDifferentOrganisation(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "a-different-org-id"})

	donorStore := &donorStore{}

	err := donorStore.DeleteDonorAccess(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")})
	assert.Error(t, err)
}

func TestDonorStoreDeleteDonorAccessWhenOneByPartialSKError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "org-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), actor.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("donor-sub")}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.DeleteDonorAccess(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")})
	assert.Error(t, err)
}

func TestDonorStoreDeleteDonorAccessWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "org-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), actor.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("donor-sub")}, nil)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.DeleteDonorAccess(ctx, sharecode.Data{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")), LpaKey: dynamo.LpaKey("lpa-id")})
	assert.Error(t, err)
}