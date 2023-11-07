package app

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func (m *mockDynamoClient) ExpectOne(ctx, pk, sk, data interface{}, err error) {
	m.
		On("One", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectOneByPartialSk(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("OneByPartialSk", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByPartialSk(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("AllByPartialSk", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllForActor(ctx, sk, data interface{}, err error) {
	m.
		On("AllForActor", ctx, sk, mock.Anything).
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

func TestDonorStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSk(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestDonorStoreGetAnyWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetAnyWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneByPartialSk(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, "LPA#an-id", "#DONOR#456", &page.Lpa{ID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestDonorStoreGetWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOne(ctx, "LPA#an-id", "#DONOR#456", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreLatest(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, "#DONOR#456", &page.Lpa{ID: "an-id"}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	lpa, err := donorStore.Latest(ctx)
	assert.Nil(t, err)
	assert.Equal(t, &page.Lpa{ID: "an-id"}, lpa)
}

func TestDonorStoreLatestWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreLatestWhenDataStoreError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectLatestForActor(ctx, "#DONOR#456", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Latest(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	testcases := map[string]struct {
		input, saved *page.Lpa
	}{
		"no uid": {
			input: &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5", HasSentApplicationUpdatedEvent: true},
			saved: &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5", HasSentApplicationUpdatedEvent: true},
		},
		"with uid": {
			input: &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5", HasSentApplicationUpdatedEvent: true, UID: "M"},
			saved: &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5", HasSentApplicationUpdatedEvent: true, UID: "M", UpdatedAt: now},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				On("Put", ctx, tc.saved).
				Return(nil)

			donorStore := &donorStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

			err := donorStore.Put(ctx, tc.input)
			assert.Nil(t, err)
		})
	}
}

func TestDonorStorePutWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.On("Put", ctx, mock.Anything).Return(expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: time.Now}

	err := donorStore.Put(ctx, &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5"})
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenUIDNeeded(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendUidRequested", ctx, event.UidRequested{
			LpaID:          "5",
			DonorSessionID: "an-id",
			Type:           "hw",
			Donor: uid.DonorDetails{
				Name:     "John Smith",
				Dob:      date.New("2000", "01", "01"),
				Postcode: "F1 1FF",
			},
		}).
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK: "LPA#5",
			SK: "#DONOR#an-id",
			ID: "5",
			Donor: actor.Donor{
				FirstNames:  "John",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "01", "01"),
				Address: place.Address{
					Line1:    "line",
					Postcode: "F1 1FF",
				},
			},
			Type:                     page.LpaTypeHealthWelfare,
			HasSentUidRequestedEvent: true,
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK: "LPA#5",
		SK: "#DONOR#an-id",
		ID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Line1:    "line",
				Postcode: "F1 1FF",
			},
		},
		Type: page.LpaTypeHealthWelfare,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDNeededMissingSessionData(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{}

	err := donorStore.Put(ctx, &page.Lpa{
		PK: "LPA#5",
		SK: "#DONOR#an-id",
		ID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Line1:    "line",
				Postcode: "F1 1FF",
			},
		},
		Type: page.LpaTypeHealthWelfare,
	})

	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStorePutWhenUIDFails(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendUidRequested", ctx, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{eventClient: eventClient, now: time.Now}

	err := donorStore.Put(ctx, &page.Lpa{
		PK: "LPA#5",
		SK: "#DONOR#an-id",
		ID: "5",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Postcode: "F1 1FF",
			},
		},
		Type: page.LpaTypeHealthWelfare,
	})

	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenApplicationUpdatedWhenError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendApplicationUpdated", ctx, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:  "LPA#5",
		SK:  "#DONOR#an-id",
		ID:  "5",
		UID: "M-1111",
		Donor: actor.Donor{
			FirstNames:  "John",
			LastName:    "Smith",
			DateOfBirth: date.New("2000", "01", "01"),
			Address: place.Address{
				Postcode: "F1 1FF",
			},
		},
		Type: page.LpaTypeHealthWelfare,
	})

	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenPreviousApplicationLinked(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendPreviousApplicationLinked", ctx, event.PreviousApplicationLinked{
			UID:                       "M-1111",
			PreviousApplicationNumber: "5555",
		}).
		Return(nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK:                                    "LPA#5",
			SK:                                    "#DONOR#an-id",
			ID:                                    "5",
			UID:                                   "M-1111",
			UpdatedAt:                             now,
			PreviousApplicationNumber:             "5555",
			HasSentApplicationUpdatedEvent:        true,
			HasSentPreviousApplicationLinkedEvent: true,
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		PreviousApplicationNumber:      "5555",
		HasSentApplicationUpdatedEvent: true,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenPreviousApplicationLinkedWontResend(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                                    "LPA#5",
		SK:                                    "#DONOR#an-id",
		ID:                                    "5",
		UID:                                   "M-1111",
		PreviousApplicationNumber:             "5555",
		HasSentApplicationUpdatedEvent:        true,
		HasSentPreviousApplicationLinkedEvent: true,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenPreviousApplicationLinkedWhenError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendPreviousApplicationLinked", ctx, mock.Anything).
		Return(expectedError)

	donorStore := &donorStore{eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		PreviousApplicationNumber:      "5555",
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenReducedFeeRequestedAndUnsentDocuments(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK:                             "LPA#5",
			SK:                             "#DONOR#an-id",
			ID:                             "5",
			UID:                            "M-1111",
			UpdatedAt:                      now,
			FeeType:                        pay.HalfFee,
			Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
			HasSentApplicationUpdatedEvent: true,
			EvidenceDelivery:               pay.Upload,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendReducedFeeRequested", ctx, event.ReducedFeeRequested{
			UID:              "M-1111",
			RequestType:      "HalfFee",
			Evidence:         []string{"lpa-uid-evidence-a-uid"},
			EvidenceDelivery: "upload",
		}).
		Return(nil)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", ctx).
		Return(page.Documents{
			{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf", Scanned: true},
		}, nil)
	documentStore.
		On("BatchPut", ctx, []page.Document{{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf", Scanned: true, Sent: now}}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, now: func() time.Time { return now }, documentStore: documentStore}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		FeeType:                        pay.HalfFee,
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
		EvidenceDelivery:               pay.Upload,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenReducedFeeRequestedWontResend(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", ctx).
		Return(page.Documents{
			{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf", Sent: now},
		}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: func() time.Time { return now }, documentStore: documentStore}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenReducedFeeRequestedWhenDocumentStoreGetAllError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", ctx).
		Return(page.Documents{}, expectedError)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	donorStore := &donorStore{now: func() time.Time { return now }, documentStore: documentStore, logger: logger, dynamoClient: dynamoClient}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
	})

	assert.Nil(t, err)
}

func TestDonorStorePutWhenReducedFeeRequestedAndUnsentDocumentsWhenEventClientSendError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendReducedFeeRequested", ctx, event.ReducedFeeRequested{
			UID:              "M-1111",
			RequestType:      "HalfFee",
			Evidence:         []string{"lpa-uid-evidence-a-uid"},
			EvidenceDelivery: "upload",
		}).
		Return(expectedError)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", ctx).
		Return(page.Documents{
			{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf", Scanned: true},
		}, nil)

	donorStore := &donorStore{eventClient: eventClient, now: func() time.Time { return now }, documentStore: documentStore}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		FeeType:                        pay.HalfFee,
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
		EvidenceDelivery:               pay.Upload,
	})

	assert.Equal(t, expectedError, err)
}

func TestDonorStorePutWhenReducedFeeRequestedAndUnsentDocumentsWhenDocumentStoreBatchPutError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("SendReducedFeeRequested", ctx, event.ReducedFeeRequested{
			UID:              "M-1111",
			RequestType:      "HalfFee",
			Evidence:         []string{"lpa-uid-evidence-a-uid"},
			EvidenceDelivery: "upload",
		}).
		Return(nil)

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("GetAll", ctx).
		Return(page.Documents{
			{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf", Scanned: true},
		}, nil)
	documentStore.
		On("BatchPut", ctx, mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{eventClient: eventClient, now: func() time.Time { return now }, documentStore: documentStore, dynamoClient: dynamoClient, logger: logger}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		FeeType:                        pay.HalfFee,
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
		EvidenceDelivery:               pay.Upload,
	})

	assert.Nil(t, err)
}

func TestDonorStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	now := time.Now()
	lpa := &page.Lpa{PK: "LPA#10100000", SK: "#DONOR#an-id", ID: "10100000", CreatedAt: now, Version: 1}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", ctx, lpa).
		Return(nil)
	dynamoClient.
		On("Create", ctx, lpaLink{PK: "LPA#10100000", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor, UpdatedAt: now}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }, now: func() time.Time { return now }}

	result, err := donorStore.Create(ctx)
	assert.Nil(t, err)
	assert.Equal(t, lpa, result)
}

func TestDonorStoreCreateWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }, now: func() time.Time { return time.Now() }}

	_, err := donorStore.Create(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreCreateWhenError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	now := time.Now()

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"certificate provider record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				On("Create", ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"link record": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				On("Create", ctx, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.
				On("Create", ctx, mock.Anything).
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
				now:          func() time.Time { return now },
			}

			_, err := donorStore.Create(ctx)
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
	dynamoClient.
		On("AllKeysByPk", ctx, "LPA#123").
		Return(keys, nil)
	dynamoClient.
		On("DeleteKeys", ctx, keys).
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
	dynamoClient.
		On("AllKeysByPk", ctx, "LPA#123").
		Return(keys, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.NotNil(t, err)
}

func TestDonorStoreDeleteWhenAllKeysByPkErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllKeysByPk", ctx, "LPA#123").
		Return(nil, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient}

	err := donorStore.Delete(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreDeleteWhenDeleteKeysErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", LpaID: "123"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("AllKeysByPk", ctx, "LPA#123").
		Return([]dynamo.Key{{PK: "LPA#123", SK: "#DONOR#an-id"}}, nil)
	dynamoClient.
		On("DeleteKeys", ctx, mock.Anything).
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
