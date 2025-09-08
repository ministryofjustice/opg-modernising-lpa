package donor

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gohugoio/hashstructure"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDonorStoreGetAny(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey(""), &donordata.Provided{LpaID: "an-id"}, nil)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &donordata.Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWhenOrganisation(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", OrganisationID: "x"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.OrganisationKey(""), &donordata.Provided{LpaID: "an-id"}, nil)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &donordata.Provided{LpaID: "an-id"}, lpa)
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
		&donordata.Provided{LpaID: "an-id"}, nil)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &donordata.Provided{SK: dynamo.LpaOwnerKey(dynamo.DonorKey("donor")), LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWithSessionMissing(t *testing.T) {
	donorStore := &Store{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreGetAnyWhenDataStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey(""), &donordata.Provided{LpaID: "an-id"}, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

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

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

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
			dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), tc.expectedSK, &donordata.Provided{LpaID: "an-id"}, nil)

			donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

			lpa, err := donorStore.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, &donordata.Provided{LpaID: "an-id"}, lpa)
		})
	}
}

func TestDonorStoreGetWhenReferenced(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey("456"), lpaReference{ReferencedSK: dynamo.OrganisationKey("789")}, nil)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.OrganisationKey("789"), &donordata.Provided{LpaID: "an-id"}, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, &donordata.Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetWithSessionMissing(t *testing.T) {
	donorStore := &Store{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreGetWhenDataStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("an-id"), dynamo.DonorKey("456"), lpaReference{ReferencedSK: "ref"}, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreLatest(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, dynamo.DonorKey("456"), &donordata.Provided{LpaID: "an-id"}, nil)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Latest(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &donordata.Provided{LpaID: "an-id"}, lpa)
}

func TestDonorStoreLatestWithSessionMissing(t *testing.T) {
	donorStore := &Store{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreLatestWhenDataStoreError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, dynamo.DonorKey("456"), &donordata.Provided{LpaID: "an-id"}, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGetByKeys(t *testing.T) {
	keys := []dynamo.Keys{
		{PK: dynamo.LpaKey("1"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a"))},
		{PK: dynamo.LpaKey("2"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("b"))},
		{PK: dynamo.LpaKey("3"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("c"))},
	}
	donors := []donordata.Provided{
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

	donorStore := &Store{dynamoClient: dynamoClient}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Equal(t, donors, result)
}

func TestDonorStoreGetByKeysWhenMissingResults(t *testing.T) {
	keys := []dynamo.Keys{
		{PK: dynamo.LpaKey("1"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a"))},
		{PK: dynamo.LpaKey("2"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("b"))},
		{PK: dynamo.LpaKey("3"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("c"))},
	}
	donors := []donordata.Provided{
		{PK: dynamo.LpaKey("1"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a")), LpaID: "1"},
		{PK: dynamo.LpaKey("3"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("c")), LpaID: "3"},
	}
	av0, _ := attributevalue.MarshalMap(donors[1])
	av1, _ := attributevalue.MarshalMap(donors[0])

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByKeys(ctx, keys).
		Return([]map[string]types.AttributeValue{av0, av1}, nil)

	donorStore := &Store{dynamoClient: dynamoClient}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Equal(t, donors, result)
}

func TestDonorStoreGetByKeysWhenNoKeys(t *testing.T) {
	keys := []dynamo.Keys{}

	donorStore := &Store{}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestDonorStoreGetByKeysWhenDynamoErrors(t *testing.T) {
	keys := []dynamo.Keys{{}}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByKeys(ctx, keys,
		nil, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient}

	_, err := donorStore.GetByKeys(ctx, keys)
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePut(t *testing.T) {
	saved := &donordata.Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"}}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, &donordata.Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenPaper(t *testing.T) {
	saved := &donordata.Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", Donor: donordata.Donor{Channel: lpadata.ChannelPaper, FirstNames: "x", LastName: "y"}}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, &donordata.Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", Donor: donordata.Donor{Channel: lpadata.ChannelPaper, FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenDonorCanChange(t *testing.T) {
	ctx := appcontext.ContextWithData(ctx, appcontext.Data{ActorType: actor.TypeDonor})

	initial := &donordata.Provided{
		PK:                             dynamo.LpaKey("5"),
		SK:                             dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		Hash:                           5,
		LpaID:                          "5",
		HasSentApplicationUpdatedEvent: true,
		Donor:                          donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"},
	}
	initial.UpdateCheckedHash()
	initial.SignedAt = testNow

	saved := &donordata.Provided{
		PK:                             dynamo.LpaKey("5"),
		SK:                             dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		CheckedHash:                    initial.CheckedHash,
		LpaID:                          "5",
		HasSentApplicationUpdatedEvent: true,
		Donor:                          donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"},
		SignedAt:                       testNow,
	}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, initial)
	assert.Nil(t, err)
}

func TestDonorStorePutWhenDonorCannotChange(t *testing.T) {
	ctx := appcontext.ContextWithData(ctx, appcontext.Data{ActorType: actor.TypeDonor})

	initial := &donordata.Provided{
		PK:                             dynamo.LpaKey("5"),
		SK:                             dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		Hash:                           5,
		LpaID:                          "5",
		HasSentApplicationUpdatedEvent: true,
		Donor:                          donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"},
		SignedAt:                       testNow,
	}
	initial.UpdateCheckedHash()
	initial.Donor.FirstNames = "z"

	donorStore := &Store{now: testNowFn}

	err := donorStore.Put(ctx, initial)
	assert.Error(t, err)
}

func TestDonorStorePutWhenOtherActorCanChange(t *testing.T) {
	ctx := appcontext.ContextWithData(ctx, appcontext.Data{ActorType: actor.TypeAttorney})

	initial := &donordata.Provided{
		PK:                             dynamo.LpaKey("5"),
		SK:                             dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		Hash:                           5,
		LpaID:                          "5",
		HasSentApplicationUpdatedEvent: true,
		Donor:                          donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"},
		SignedAt:                       testNow,
	}
	initial.UpdateCheckedHash()
	initial.Donor.FirstNames = "z"

	saved := &donordata.Provided{
		PK:                             dynamo.LpaKey("5"),
		SK:                             dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		CheckedHash:                    initial.CheckedHash,
		LpaID:                          "5",
		HasSentApplicationUpdatedEvent: true,
		Donor:                          donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "z", LastName: "y"},
		SignedAt:                       testNow,
	}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, initial)
	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDSet(t *testing.T) {
	saved := &donordata.Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", UpdatedAt: testNow, Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"}}
	saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, search.Lpa{PK: dynamo.LpaKey("5").PK(), SK: dynamo.DonorKey("an-id").SK(), Donor: search.LpaDonor{FirstNames: "x", LastName: "y"}}).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &donordata.Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"}})
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

	donorStore := &Store{dynamoClient: dynamoClient, searchClient: searchClient, logger: logger, now: testNowFn}

	err := donorStore.Put(ctx, &donordata.Provided{PK: dynamo.LpaKey("5"), Hash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenNoChange(t *testing.T) {
	donorStore := &Store{}

	donor := &donordata.Provided{LpaID: "an-id"}
	donor.Hash, _ = hashstructure.Hash(donor, nil)

	err := donorStore.Put(ctx, donor)
	assert.Nil(t, err)
}

func TestDonorStorePutWhenCheckChangeAndCheckCompleted(t *testing.T) {
	saved := &donordata.Provided{PK: dynamo.LpaKey("5"), Hash: 5, CheckedHash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "a", LastName: "b"}, Tasks: donordata.Tasks{CheckYourLpa: task.StateInProgress}}
	_ = saved.UpdateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, &donordata.Provided{PK: dynamo.LpaKey("5"), Hash: 5, CheckedHash: 5, SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: donordata.Donor{Channel: lpadata.ChannelOnline, FirstNames: "a", LastName: "b"}, Tasks: donordata.Tasks{CheckYourLpa: task.StateCompleted}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().Put(ctx, mock.Anything).Return(expectedError)

	donorStore := &Store{dynamoClient: dynamoClient, now: time.Now}

	err := donorStore.Put(ctx, &donordata.Provided{PK: dynamo.LpaKey("5"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaID: "5"})
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

	donorStore := &Store{eventClient: eventClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &donordata.Provided{
		PK:     dynamo.LpaKey("5"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
		LpaID:  "5",
		LpaUID: "M-1111",
		Donor: donordata.Donor{
			Channel:     lpadata.ChannelOnline,
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
	testCases := map[string]donordata.Provided{
		"with previous details": {
			Donor: donordata.Donor{
				UID:                  actoruid.New(),
				FirstNames:           "a",
				LastName:             "b",
				OtherNames:           "c",
				DateOfBirth:          date.New("2000", "01", "02"),
				Address:              place.Address{Line1: "d"},
				InternationalAddress: place.InternationalAddress{Town: "a-town"},
			},
		},
		"no previous details": {},
	}

	for name, previousDetails := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})
			donor := &donordata.Provided{
				PK:        dynamo.LpaKey("10100000"),
				SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
				LpaID:     "10100000",
				CreatedAt: testNow,
				Version:   1,
				Donor: donordata.Donor{
					UID:                  testUID,
					FirstNames:           previousDetails.Donor.FirstNames,
					LastName:             previousDetails.Donor.LastName,
					OtherNames:           previousDetails.Donor.OtherNames,
					DateOfBirth:          previousDetails.Donor.DateOfBirth,
					Address:              previousDetails.Donor.Address,
					InternationalAddress: previousDetails.Donor.InternationalAddress,
					Channel:              lpadata.ChannelOnline,
				},
			}
			donor.UpdateHash()

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), previousDetails, nil)
			dynamoClient.EXPECT().
				WriteTransaction(ctx, &dynamo.Transaction{
					Creates: []any{
						dynamo.Keys{PK: dynamo.LpaKey("10100000"), SK: dynamo.ReservedKey(dynamo.DonorKey)},
						donor,
						dashboarddata.LpaLink{
							PK:        dynamo.LpaKey("10100000"),
							SK:        dynamo.SubKey("an-id"),
							DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
							UID:       donor.Donor.UID,
							ActorType: actor.TypeDonor,
							UpdatedAt: testNow,
						},
					},
				}).
				Return(nil)

			donorStore := &Store{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }, now: testNowFn, newUID: testUIDFn}

			result, err := donorStore.Create(ctx)
			assert.Nil(t, err)
			assert.Equal(t, donor, result)
		})
	}
}

func TestDonorStoreCreateWithSessionMissing(t *testing.T) {
	donorStore := &Store{dynamoClient: nil, uuidString: func() string { return "10100000" }, now: func() time.Time { return time.Now() }}

	_, err := donorStore.Create(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreCreateWhenError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"latest": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), donordata.Provided{}, expectedError)

			return dynamoClient
		},
		"transaction": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, dynamo.DonorKey("an-id"), donordata.Provided{}, nil)
			dynamoClient.EXPECT().
				WriteTransaction(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDataStore := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDataStore(t)

			donorStore := &Store{
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
	testcases := map[string][]dashboarddata.LpaLink{
		"no link": {},
		"not a donor link": {{
			PK:        dynamo.LpaKey(""),
			SK:        dynamo.SubKey("a-sub"),
			ActorType: actor.TypeCertificateProvider,
		}},
	}

	for name, links := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
			accessCode := accesscodedata.Link{
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
					dashboarddata.LpaLink{
						PK:        dynamo.LpaKey("lpa-id"),
						SK:        dynamo.SubKey("session-id"),
						DonorKey:  dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
						ActorType: actor.TypeDonor,
						UpdatedAt: testNowFn(),
					},
				},
				Updates: []*types.Update{
					{
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: dynamo.LpaKey("lpa-id").PK()},
							"SK": &types.AttributeValueMemberS{Value: dynamo.OrganisationLinkKey("org-id").SK()},
						},
						UpdateExpression: aws.String("SET #Field1 = :Value1, #Field2 = :Value2"),
						ExpressionAttributeValues: map[string]types.AttributeValue{
							":Value1": &types.AttributeValueMemberS{Value: "a@example.com"},
							":Value2": &types.AttributeValueMemberS{Value: testNow.Format(time.RFC3339Nano)},
						},
						ExpressionAttributeNames: map[string]string{
							"#Field1": "LpaLinkedTo",
							"#Field2": "LpaLinkedAt",
						},
					},
				},
				Deletes: []dynamo.Keys{
					{PK: accessCode.PK, SK: accessCode.SK},
				},
			}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				AllByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), mock.Anything).
				Return(nil).
				SetData(links)
			dynamoClient.EXPECT().
				WriteTransaction(ctx, expectedTransaction).
				Return(nil)

			donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			err := donorStore.Link(ctx, accessCode, "a@example.com")
			assert.Nil(t, err)
		})
	}
}

func TestDonorStoreLinkWithDonor(t *testing.T) {
	donorStore := &Store{}

	err := donorStore.Link(ctx, accesscodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("donor"))}, "a@example.com")
	assert.Error(t, err)
}

func TestDonorStoreLinkWithSessionMissing(t *testing.T) {
	donorStore := &Store{}

	err := donorStore.Link(ctx, accesscodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))}, "a@example.com")
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreLinkWithSessionIDMissing(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})
	donorStore := &Store{}

	err := donorStore.Link(ctx, accesscodedata.Link{LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))}, "a@example.com")
	assert.Error(t, err)
}

func TestDonorStoreLinkWhenDonorLinkAlreadyExists(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), mock.Anything).
		Return(nil).
		SetData([]dashboarddata.LpaLink{
			{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("a-sub"), ActorType: actor.TypeDonor},
		})

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.Link(
		ctx,
		accesscodedata.Link{LpaKey: dynamo.LpaKey("lpa-id"), LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org"))},
		"a@example.com",
	)

	assert.Equal(t, errors.New("a donor link already exists for lpa-id"), err)
}

func TestDonorStoreLinkWhenError(t *testing.T) {
	testcases := map[string]func(*mockDynamoClient){
		"AllByPartialSK errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.EXPECT().
				AllByPartialSK(mock.Anything, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), mock.Anything).
				Return(expectedError)
		},
		"WriteTransaction errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.EXPECT().
				AllByPartialSK(mock.Anything, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), mock.Anything).
				Return(nil).
				SetData([]dashboarddata.LpaLink{
					{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("a-sub"), ActorType: actor.TypeAttorney},
				})
			dynamoClient.EXPECT().
				WriteTransaction(mock.Anything, mock.Anything).
				Return(expectedError)
		},
	}

	for name, setupDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"})
			accessCode := accesscodedata.Link{
				LpaKey:      dynamo.LpaKey("lpa-id"),
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
			}

			dynamoClient := newMockDynamoClient(t)
			setupDynamoClient(dynamoClient)

			donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

			err := donorStore.Link(ctx, accessCode, "a@example.com")
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
	dynamoClient.ExpectOne(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("an-id"),
		&donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), LpaUID: "lpa-uid"}, nil)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, keys).
		Return(nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Delete(ctx, search.Lpa{PK: dynamo.LpaKey("123").PK(), SK: dynamo.DonorKey("an-id").SK()}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationDeleted(ctx, event.ApplicationDeleted{UID: "lpa-uid"}).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient, eventClient: eventClient, searchClient: searchClient}

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

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.NotNil(t, err)
}

func TestDonorStoreDeleteWhenErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "123"})

	testCases := map[string]struct {
		dynamoClient func(t *testing.T) *mockDynamoClient
		eventClient  func(t *testing.T) *mockEventClient
		searchClient func(t *testing.T) *mockSearchClient
	}{
		"dynamo AllKeysByPK": {
			dynamoClient: func(t *testing.T) *mockDynamoClient {
				dynamoClient := newMockDynamoClient(t)
				dynamoClient.EXPECT().
					AllKeysByPK(mock.Anything, mock.Anything).
					Return(nil, expectedError)
				return dynamoClient
			},
			searchClient: func(t *testing.T) *mockSearchClient { return nil },
			eventClient:  func(t *testing.T) *mockEventClient { return nil },
		},
		"dynamo One": {
			dynamoClient: func(t *testing.T) *mockDynamoClient {
				dynamoClient := newMockDynamoClient(t)
				dynamoClient.EXPECT().
					AllKeysByPK(mock.Anything, mock.Anything).
					Return([]dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")}}, nil)
				dynamoClient.ExpectOne(mock.Anything, mock.Anything, mock.Anything,
					&donordata.Provided{}, expectedError)
				return dynamoClient
			},
			searchClient: func(t *testing.T) *mockSearchClient { return nil },
			eventClient:  func(t *testing.T) *mockEventClient { return nil },
		},
		"search delete": {
			dynamoClient: func(t *testing.T) *mockDynamoClient {
				dynamoClient := newMockDynamoClient(t)
				dynamoClient.EXPECT().
					AllKeysByPK(mock.Anything, mock.Anything).
					Return([]dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")}}, nil)
				dynamoClient.ExpectOne(mock.Anything, mock.Anything, mock.Anything,
					&donordata.Provided{}, nil)
				return dynamoClient
			},
			searchClient: func(t *testing.T) *mockSearchClient {
				searchClient := newMockSearchClient(t)
				searchClient.EXPECT().
					Delete(ctx, mock.Anything).
					Return(expectedError)
				return searchClient
			},
			eventClient: func(t *testing.T) *mockEventClient {
				return nil
			},
		},
		"event send": {
			dynamoClient: func(t *testing.T) *mockDynamoClient {
				dynamoClient := newMockDynamoClient(t)
				dynamoClient.EXPECT().
					AllKeysByPK(mock.Anything, mock.Anything).
					Return([]dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")}}, nil)
				dynamoClient.ExpectOne(mock.Anything, mock.Anything, mock.Anything,
					&donordata.Provided{}, nil)
				return dynamoClient
			},
			searchClient: func(t *testing.T) *mockSearchClient {
				searchClient := newMockSearchClient(t)
				searchClient.EXPECT().
					Delete(ctx, mock.Anything).
					Return(nil)
				return searchClient
			},
			eventClient: func(t *testing.T) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendApplicationDeleted(ctx, mock.Anything).
					Return(expectedError)
				return eventClient
			},
		},
		"dynamo DeleteKeys": {
			dynamoClient: func(t *testing.T) *mockDynamoClient {
				dynamoClient := newMockDynamoClient(t)
				dynamoClient.EXPECT().
					AllKeysByPK(mock.Anything, mock.Anything).
					Return([]dynamo.Keys{{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")}}, nil)
				dynamoClient.ExpectOne(mock.Anything, mock.Anything, mock.Anything,
					&donordata.Provided{}, nil)
				dynamoClient.EXPECT().
					DeleteKeys(ctx, mock.Anything).
					Return(expectedError)
				return dynamoClient
			},
			searchClient: func(t *testing.T) *mockSearchClient {
				searchClient := newMockSearchClient(t)
				searchClient.EXPECT().
					Delete(ctx, mock.Anything).
					Return(nil)
				return searchClient
			},
			eventClient: func(t *testing.T) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendApplicationDeleted(mock.Anything, mock.Anything).
					Return(nil)
				return eventClient
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			donorStore := &Store{
				dynamoClient: tc.dynamoClient(t),
				eventClient:  tc.eventClient(t),
				searchClient: tc.searchClient(t),
			}

			err := donorStore.Delete(ctx)
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestDonorStoreDeleteWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"missing":      context.Background(),
		"no LpaID":     appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id"}),
		"no SessionID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{LpaID: "123"}),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			donorStore := &Store{}

			err := donorStore.Delete(ctx)
			assert.NotNil(t, err)
		})
	}
}

func TestDonorStoreDeleteDonorAccess(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "org-id"})

	link := dashboarddata.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("donor-sub"), ActorType: actor.TypeDonor}
	supporterLink := supporterdata.LpaLink{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.OrganisationLinkKey("org-id")}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.SubKey(""), mock.Anything).
		Return(nil).
		SetData([]dashboarddata.LpaLink{
			{PK: dynamo.LpaKey("no"), SK: dynamo.SubKey("no"), ActorType: actor.TypeCertificateProvider},
			link,
		})
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: link.PK, SK: link.SK},
				{PK: supporterLink.PK, SK: dynamo.DonorKey(link.UserSub())},
				{PK: supporterLink.PK, SK: supporterLink.SK},
			},
		}).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteDonorAccess(ctx, supporterLink)
	assert.Nil(t, err)
}

func TestDonorStoreDeleteDonorAccessWhenOneByPartialSKError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "org-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByPartialSK(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteDonorAccess(ctx, supporterdata.LpaLink{})
	assert.Error(t, err)
}

func TestDonorStoreDeleteDonorAccessWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", OrganisationID: "org-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllByPartialSK(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData([]dashboarddata.LpaLink{
			{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.SubKey("donor-sub"), ActorType: actor.TypeDonor},
		})
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteDonorAccess(ctx, supporterdata.LpaLink{})
	assert.Error(t, err)
}

func TestDonorStoreDeleteVoucher(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "lpa-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id")),
			accesscodedata.Link{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}, nil)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}},
			Puts: []any{
				&donordata.Provided{},
			},
		}).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteVoucher(ctx, &donordata.Provided{
		Voucher:          donordata.Voucher{FirstNames: "a"},
		VoucherInvitedAt: testNow,
		WantVoucher:      form.Yes,
	})
	assert.Nil(t, err)
}

func TestDonorStoreDeleteVoucherWhenSessionMissing(t *testing.T) {
	donorStore := &Store{}

	err := donorStore.DeleteVoucher(context.Background(), &donordata.Provided{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestDonorStoreDeleteVoucherWhenOneBySKErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "lpa-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id")),
			accesscodedata.Link{}, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteVoucher(ctx, &donordata.Provided{
		Voucher: donordata.Voucher{FirstNames: "a"},
	})
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreDeleteVoucherWhenWriteTransactionError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "lpa-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id")),
			accesscodedata.Link{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}, nil)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}},
			Puts: []any{
				&donordata.Provided{},
			},
		}).
		Return(expectedError)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteVoucher(ctx, &donordata.Provided{
		Voucher:          donordata.Voucher{FirstNames: "a"},
		VoucherInvitedAt: testNow,
		WantVoucher:      form.Yes,
	})
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreDeleteVoucherWhenAccessCodeUsed(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "lpa-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id")),
			accesscodedata.Link{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}, dynamo.NotFoundError{})
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.VoucherKey(""),
			voucherdata.Provided{
				PK: dynamo.LpaKey("lpa-id"),
				SK: dynamo.VoucherKey("voucher-id"),
			}, nil)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.VoucherKey("voucher-id")},
				{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.ReservedKey(dynamo.VoucherKey)},
			},
			Puts: []any{
				&donordata.Provided{PK: dynamo.LpaKey("lpa-id")},
			},
		}).
		Return(nil)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteVoucher(ctx, &donordata.Provided{
		PK:               dynamo.LpaKey("lpa-id"),
		Voucher:          donordata.Voucher{FirstNames: "a"},
		VoucherInvitedAt: testNow,
		WantVoucher:      form.Yes,
	})
	assert.Nil(t, err)
}

func TestDonorStoreDeleteVoucherWhenExpectOneByPartialSKError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "lpa-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id")),
			accesscodedata.Link{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}, dynamo.NotFoundError{})
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.VoucherKey(""),
			voucherdata.Provided{
				PK: dynamo.LpaKey("lpa-id"),
				SK: dynamo.VoucherKey("voucher-id"),
			}, expectedError)

	donorStore := &Store{dynamoClient: dynamoClient}

	err := donorStore.DeleteVoucher(ctx, &donordata.Provided{
		PK:               dynamo.LpaKey("lpa-id"),
		Voucher:          donordata.Voucher{FirstNames: "a"},
		VoucherInvitedAt: testNow,
		WantVoucher:      form.Yes,
	})
	assert.Equal(t, expectedError, err)
}

func TestDonorFailVoucher(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "an-id", LpaID: "lpa-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id")),
			accesscodedata.Link{
				PK: dynamo.AccessKey(dynamo.VoucherAccessKey("hey")),
				SK: dynamo.AccessSortKey(dynamo.VoucherAccessSortKey(dynamo.LpaKey("lpa-id"))),
			}, dynamo.NotFoundError{})
	dynamoClient.
		ExpectOneByPartialSK(ctx, dynamo.LpaKey("lpa-id"), dynamo.VoucherKey(""),
			voucherdata.Provided{
				PK: dynamo.LpaKey("lpa-id"),
				SK: dynamo.VoucherKey("voucher-id"),
			}, nil)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Deletes: []dynamo.Keys{
				{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.VoucherKey("voucher-id")},
				{PK: dynamo.LpaKey("lpa-id"), SK: dynamo.ReservedKey(dynamo.VoucherKey)},
			},
			Puts: []any{
				&donordata.Provided{
					PK:            dynamo.LpaKey("lpa-id"),
					WantVoucher:   form.YesNoUnknown,
					FailedVoucher: donordata.Voucher{FirstNames: "a", LastName: "b", FailedAt: testNow},
				},
			},
		}).
		Return(expectedError)

	donorStore := &Store{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.FailVoucher(ctx, &donordata.Provided{
		PK:                       dynamo.LpaKey("lpa-id"),
		Voucher:                  donordata.Voucher{FirstNames: "a", LastName: "b"},
		WantVoucher:              form.Yes,
		VoucherInvitedAt:         testNow,
		DetailsVerifiedByVoucher: true,
	})
	assert.Equal(t, expectedError, err)
}
