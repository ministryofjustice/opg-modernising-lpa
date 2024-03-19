package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/mitchellh/hashstructure/v2"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var (
	ctx           = context.Background()
	expectedError = errors.New("err")
	testNow       = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn     = func() time.Time { return testNow }
	testUID       = actoruid.New()
	testUIDFn     = func() actoruid.UID { return testUID }
)

func (m *mockDynamoClient) ExpectOne(ctx, pk, sk, data interface{}, err error) {
	m.
		On("One", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (m *mockDynamoClient) ExpectOneByPK(ctx, pk, data interface{}, err error) {
	m.
		On("OneByPK", ctx, pk, mock.Anything).
		Return(func(ctx context.Context, pk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (m *mockDynamoClient) ExpectOneByPartialSK(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("OneByPartialSK", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByPartialSK(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("AllByPartialSK", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllBySK(ctx, sk, data interface{}, err error) {
	m.
		On("AllBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, pk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectLatestForActor(ctx, sk, data interface{}, err error) {
	m.
		On("LatestForActor", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByKeys(ctx context.Context, keys []dynamo.Key, data interface{}, err error) {
	m.
		On("AllByKeys", ctx, keys, mock.Anything).
		Return(data, err)
}

func (m *mockDynamoClient) ExpectOneBySK(ctx, sk, data interface{}, err error) {
	m.
		On("OneBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func TestDonorStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, "LPA#an-id", "#DONOR#", &actor.DonorProvidedDetails{LpaID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.DonorProvidedDetails{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetAnyWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSK(ctx, "LPA#an-id", "#DONOR#", &actor.DonorProvidedDetails{LpaID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGet(t *testing.T) {
	testCases := map[string]struct {
		sessionData *page.SessionData
		expectedSK  string
	}{
		"donor": {
			sessionData: &page.SessionData{LpaID: "an-id", SessionID: "456"},
			expectedSK:  "#DONOR#456",
		},
		"organisation": {
			sessionData: &page.SessionData{LpaID: "an-id", SessionID: "456", OrganisationID: "789"},
			expectedSK:  "ORGANISATION#789",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), tc.sessionData)

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectOne(ctx, "LPA#an-id", tc.expectedSK, &actor.DonorProvidedDetails{LpaID: "an-id"}, nil)

			donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

			lpa, err := donorStore.Get(ctx)
			assert.Nil(t, err)
			assert.Equal(t, &actor.DonorProvidedDetails{LpaID: "an-id"}, lpa)
		})
	}
}

func TestDonorStoreGetWhenReferenced(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, "LPA#an-id", "#DONOR#456", lpaReference{ReferencedSK: "ORGANISATION#789"}, nil)
	dynamoClient.ExpectOne(ctx, "LPA#an-id", "ORGANISATION#789", &actor.DonorProvidedDetails{LpaID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, &actor.DonorProvidedDetails{LpaID: "an-id"}, lpa)
}

func TestDonorStoreGetWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, "LPA#an-id", "#DONOR#456", lpaReference{ReferencedSK: "ref"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreLatest(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, "#DONOR#456", &actor.DonorProvidedDetails{LpaID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Latest(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &actor.DonorProvidedDetails{LpaID: "an-id"}, lpa)
}

func TestDonorStoreLatestWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreLatestWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, "#DONOR#456", &actor.DonorProvidedDetails{LpaID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGetByKeys(t *testing.T) {
	keys := []dynamo.Key{{}}
	donors := []actor.DonorProvidedDetails{{LpaID: "1"}, {LpaID: "2"}}
	av0, _ := attributevalue.MarshalMap(donors[0])
	av1, _ := attributevalue.MarshalMap(donors[1])

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByKeys(ctx, keys,
		[]map[string]types.AttributeValue{av0, av1}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Equal(t, donors, result)
}

func TestDonorStoreGetByKeysWhenNoKeys(t *testing.T) {
	keys := []dynamo.Key{}

	donorStore := &donorStore{}

	result, err := donorStore.GetByKeys(ctx, keys)
	assert.Nil(t, err)
	assert.Empty(t, result)
}

func TestDonorStoreGetByKeysWhenDynamoErrors(t *testing.T) {
	keys := []dynamo.Key{{}}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByKeys(ctx, keys,
		nil, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	_, err := donorStore.GetByKeys(ctx, keys)
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePut(t *testing.T) {
	saved := &actor.DonorProvidedDetails{PK: "LPA#5", SK: "#DONOR#an-id", LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: actor.Donor{FirstNames: "x", LastName: "y"}}
	saved.Hash, _ = saved.GenerateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{PK: "LPA#5", Hash: 5, SK: "#DONOR#an-id", LpaID: "5", HasSentApplicationUpdatedEvent: true, Donor: actor.Donor{FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDSet(t *testing.T) {
	saved := &actor.DonorProvidedDetails{PK: "LPA#5", SK: "#DONOR#an-id", LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", UpdatedAt: testNow, Donor: actor.Donor{FirstNames: "x", LastName: "y"}}
	saved.Hash, _ = saved.GenerateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, saved).
		Return(nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, search.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", DonorFullName: "x y"}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{PK: "LPA#5", Hash: 5, SK: "#DONOR#an-id", LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", Donor: actor.Donor{FirstNames: "x", LastName: "y"}})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDSetIndexErrors(t *testing.T) {
	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{PK: "LPA#5", Hash: 5, SK: "#DONOR#an-id", LpaID: "5", HasSentApplicationUpdatedEvent: true, LpaUID: "M", Donor: actor.Donor{FirstNames: "x", LastName: "y"}})
	assert.ErrorIs(t, err, expectedError)
}

func TestDonorStorePutWhenNoChange(t *testing.T) {
	donorStore := &donorStore{}

	donor := &actor.DonorProvidedDetails{LpaID: "an-id"}
	donor.Hash, _ = hashstructure.Hash(donor, hashstructure.FormatV2, nil)

	err := donorStore.Put(ctx, donor)
	assert.Nil(t, err)
}

func TestDonorStorePutWhenError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().Put(ctx, mock.Anything).Return(expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: time.Now}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{PK: "LPA#5", SK: "#DONOR#an-id", LpaID: "5"})
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenUIDNeeded(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendUidRequested(ctx, event.UidRequested{
			LpaID:          "5",
			DonorSessionID: "an-id",
			Type:           "personal-welfare",
			Donor: uid.DonorDetails{
				Name:     "John Smith",
				Dob:      date.New("2000", "01", "01"),
				Postcode: "F1 1FF",
			},
		}).
		Return(nil)

	updatedDonor := &actor.DonorProvidedDetails{
		PK:    "LPA#5",
		SK:    "#DONOR#an-id",
		LpaID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Line1:    "line",
				Postcode: "F1 1FF",
			},
		},
		Type:                     actor.LpaTypePersonalWelfare,
		HasSentUidRequestedEvent: true,
	}
	updatedDonor.Hash, _ = updatedDonor.GenerateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, updatedDonor).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:    "LPA#5",
		SK:    "#DONOR#an-id",
		LpaID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Line1:    "line",
				Postcode: "F1 1FF",
			},
		},
		Type: actor.LpaTypePersonalWelfare,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDNeededMissingSessionData(t *testing.T) {
	donorStore := &donorStore{}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:    "LPA#5",
		SK:    "#DONOR#an-id",
		LpaID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Line1:    "line",
				Postcode: "F1 1FF",
			},
		},
		Type: actor.LpaTypePersonalWelfare,
	})

	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStorePutWhenUIDFails(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendUidRequested(ctx, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{eventClient: eventClient, now: time.Now}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:    "LPA#5",
		SK:    "#DONOR#an-id",
		LpaID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Postcode: "F1 1FF",
			},
		},
		Type: actor.LpaTypePersonalWelfare,
	})

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

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:     "LPA#5",
		SK:     "#DONOR#an-id",
		LpaID:  "5",
		LpaUID: "M-1111",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Postcode: "F1 1FF",
			},
		},
		Type: actor.LpaTypePersonalWelfare,
	})

	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenPreviousApplicationLinked(t *testing.T) {
	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPreviousApplicationLinked(ctx, event.PreviousApplicationLinked{
			UID:                       "M-1111",
			PreviousApplicationNumber: "5555",
		}).
		Return(nil)

	updatedDonor := &actor.DonorProvidedDetails{
		PK:                                    "LPA#5",
		SK:                                    "#DONOR#an-id",
		LpaID:                                 "5",
		LpaUID:                                "M-1111",
		UpdatedAt:                             testNow,
		PreviousApplicationNumber:             "5555",
		HasSentApplicationUpdatedEvent:        true,
		HasSentPreviousApplicationLinkedEvent: true,
	}
	updatedDonor.Hash, _ = updatedDonor.GenerateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, updatedDonor).
		Return(nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		LpaID:                          "5",
		LpaUID:                         "M-1111",
		PreviousApplicationNumber:      "5555",
		HasSentApplicationUpdatedEvent: true,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenPreviousApplicationLinkedWontResend(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:                                    "LPA#5",
		SK:                                    "#DONOR#an-id",
		LpaID:                                 "5",
		LpaUID:                                "M-1111",
		PreviousApplicationNumber:             "5555",
		HasSentApplicationUpdatedEvent:        true,
		HasSentPreviousApplicationLinkedEvent: true,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenPreviousApplicationLinkedWhenError(t *testing.T) {
	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendPreviousApplicationLinked(ctx, mock.Anything).
		Return(expectedError)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{eventClient: eventClient, searchClient: searchClient, now: testNowFn}

	err := donorStore.Put(ctx, &actor.DonorProvidedDetails{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		LpaID:                          "5",
		LpaUID:                         "M-1111",
		PreviousApplicationNumber:      "5555",
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreCreate(t *testing.T) {
	testCases := map[string]actor.DonorProvidedDetails{
		"with previous details": {
			Donor: actor.Donor{
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
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
			donor := &actor.DonorProvidedDetails{
				PK:        "LPA#10100000",
				SK:        "#DONOR#an-id",
				LpaID:     "10100000",
				CreatedAt: testNow,
				Version:   1,
				Donor: actor.Donor{
					UID:         testUID,
					FirstNames:  previousDetails.Donor.FirstNames,
					LastName:    previousDetails.Donor.LastName,
					OtherNames:  previousDetails.Donor.OtherNames,
					DateOfBirth: previousDetails.Donor.DateOfBirth,
					Address:     previousDetails.Donor.Address,
				},
			}
			donor.Hash, _ = donor.GenerateHash()

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, "#DONOR#an-id", previousDetails, nil)
			dynamoClient.EXPECT().
				Create(ctx, donor).
				Return(nil)
			dynamoClient.EXPECT().
				Create(ctx, lpaLink{PK: "LPA#10100000", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor, UpdatedAt: testNow}).
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
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreCreateWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"latest": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, "#DONOR#an-id", actor.DonorProvidedDetails{}, expectedError)

			return dynamoClient
		},
		"donor record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, "#DONOR#an-id", actor.DonorProvidedDetails{}, nil)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"link record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectLatestForActor(ctx, "#DONOR#an-id", actor.DonorProvidedDetails{}, nil)
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
		link                lpaLink
	}{
		"no link": {
			oneByPartialSKError: dynamo.NotFoundError{},
		},
		"not a donor link": {
			link: lpaLink{
				PK:        "LPA#",
				SK:        "#SUB#a-sub",
				ActorType: actor.TypeCertificateProvider,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
			shareCode := actor.ShareCodeData{
				LpaID:     "lpa-id",
				SessionID: "org-id",
			}

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneByPartialSK(ctx, "LPA#lpa-id", "#SUB#", tc.link, tc.oneByPartialSKError)
			dynamoClient.EXPECT().
				Create(ctx, lpaReference{
					PK:           "LPA#lpa-id",
					SK:           "#DONOR#an-id",
					ReferencedSK: "ORGANISATION#org-id",
				}).
				Return(nil).
				Once()
			dynamoClient.EXPECT().
				Create(ctx, lpaLink{
					PK:        "LPA#lpa-id",
					SK:        "#SUB#an-id",
					DonorKey:  "ORGANISATION#org-id",
					ActorType: actor.TypeDonor,
					UpdatedAt: testNow,
				}).
				Return(nil).
				Once()

			donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

			err := donorStore.Link(ctx, shareCode)
			assert.Nil(t, err)
		})
	}
}

func TestDonorStoreLinkWithSessionMissing(t *testing.T) {
	donorStore := &donorStore{}

	err := donorStore.Link(ctx, actor.ShareCodeData{})
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreLinkWithSessionIDMissing(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})
	donorStore := &donorStore{}

	err := donorStore.Link(ctx, actor.ShareCodeData{})
	assert.Error(t, err)
}

func TestDonorStoreLinkWhenDonorLinkAlreadyExists(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSK(ctx, "LPA#lpa-id", "#SUB#", lpaLink{PK: "LPA#lpa-id", SK: "#SUB#a-sub", ActorType: actor.TypeDonor}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Link(ctx, actor.ShareCodeData{LpaID: "lpa-id"})
	assert.Equal(t, errors.New("a donor link already exists for lpa-id"), err)
}

func TestDonorStoreLinkWhenError(t *testing.T) {
	testcases := map[string]func(*mockDynamoClient){
		"OneByPartialSK errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.
				ExpectOneByPartialSK(mock.Anything, "LPA#lpa-id", "#SUB#", lpaLink{PK: "LPA#lpa-id", SK: "#SUB#a-sub", ActorType: actor.TypeAttorney}, expectedError)
		},
		"lpaReference errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.
				ExpectOneByPartialSK(mock.Anything, "LPA#lpa-id", "#SUB#", lpaLink{PK: "LPA#lpa-id", SK: "#SUB#a-sub", ActorType: actor.TypeAttorney}, nil)
			dynamoClient.EXPECT().
				Create(mock.Anything, mock.Anything).
				Return(expectedError)
		},
		"lpaLink errors": func(dynamoClient *mockDynamoClient) {
			dynamoClient.
				ExpectOneByPartialSK(mock.Anything, "LPA#lpa-id", "#SUB#", lpaLink{PK: "LPA#lpa-id", SK: "#SUB#a-sub", ActorType: actor.TypeAttorney}, nil)
			dynamoClient.EXPECT().
				Create(mock.Anything, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.EXPECT().
				Create(mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
	}

	for name, setupDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
			shareCode := actor.ShareCodeData{
				LpaID:     "lpa-id",
				SessionID: "org-id",
			}

			dynamoClient := newMockDynamoClient(t)
			setupDynamoClient(dynamoClient)

			donorStore := &donorStore{dynamoClient: dynamoClient, now: testNowFn}

			err := donorStore.Link(ctx, shareCode)
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestDonorStoreDelete(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	keys := []dynamo.Key{
		{PK: "LPA#123", SK: "sk1"},
		{PK: "LPA#123", SK: "sk2"},
		{PK: "LPA#123", SK: "#DONOR#an-id"},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, "LPA#123").
		Return(keys, nil)
	dynamoClient.EXPECT().
		DeleteKeys(ctx, keys).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.Nil(t, err)
}

func TestDonorStoreDeleteWhenOtherDonor(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	keys := []dynamo.Key{
		{PK: "LPA#123", SK: "sk1"},
		{PK: "LPA#123", SK: "sk2"},
		{PK: "LPA#123", SK: "#DONOR#another-id"},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, "LPA#123").
		Return(keys, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.NotNil(t, err)
}

func TestDonorStoreDeleteWhenAllKeysByPKErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, "LPA#123").
		Return(nil, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreDeleteWhenDeleteKeysErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		AllKeysByPK(ctx, "LPA#123").
		Return([]dynamo.Key{{PK: "LPA#123", SK: "#DONOR#an-id"}}, nil)
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
		"no LpaID":     page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"}),
		"no SessionID": page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "123"}),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			donorStore := &donorStore{}

			err := donorStore.Delete(ctx)
			assert.NotNil(t, err)
		})
	}
}
