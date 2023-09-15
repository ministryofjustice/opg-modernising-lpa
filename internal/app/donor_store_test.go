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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func (m *mockDynamoClient) ExpectGet(ctx, pk, sk, data interface{}, err error) {
	m.
		On("Get", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectGetOneByPartialSk(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("GetOneByPartialSk", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk, partialSk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectGetAllByGsi(ctx, gsi, sk, data interface{}, err error) {
	m.
		On("GetAllByGsi", ctx, gsi, sk, mock.Anything).
		Return(func(ctx context.Context, gsi, pk string, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectGetAllByKeys(ctx context.Context, keys []dynamo.Key, data interface{}, err error) {
	m.
		On("GetAllByKeys", ctx, keys, mock.Anything).
		Return(data, err)
}

func TestDonorStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	lpa123 := &page.Lpa{ID: "123", UpdatedAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}
	lpa456 := &page.Lpa{ID: "456", UpdatedAt: time.Date(2002, time.January, 1, 0, 0, 0, 0, time.UTC)}
	lpa789 := &page.Lpa{ID: "789", UpdatedAt: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectGetAllByGsi(ctx, "ActorIndex", "#DONOR#an-id",
		[]any{lpa123, lpa456, lpa789}, nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: nil}

	result, err := donorStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []*page.Lpa{lpa456, lpa789, lpa123}, result)
}

func TestDonorStoreGetAllWithSessionMissing(t *testing.T) {
	ctx := context.Background()

	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAll(ctx)
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestDonorStoreGetAllWhenMissingSessionID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	donorStore := &donorStore{dynamoClient: nil, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAll(ctx)
	assert.NotNil(t, err)
}

func TestDonorStoreGetAny(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectGetOneByPartialSk(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, nil)

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
	dynamoClient.ExpectGetOneByPartialSk(ctx, "LPA#an-id", "#DONOR#", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.GetAny(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{LpaID: "an-id", SessionID: "456"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectGet(ctx, "LPA#an-id", "#DONOR#456", &page.Lpa{ID: "an-id"}, nil)

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
	dynamoClient.ExpectGet(ctx, "LPA#an-id", "#DONOR#456", &page.Lpa{ID: "an-id"}, expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uuidString: func() string { return "10100000" }}

	_, err := donorStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDonorStorePut(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5", UpdatedAt: now}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{PK: "LPA#5", SK: "#DONOR#an-id", ID: "5"})
	assert.Nil(t, err)
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
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "application-updated", applicationUpdatedEvent{
			UID:  "M-1111",
			Type: "hw",
			Donor: applicationUpdatedEventDonor{
				FirstNames:  "John",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "01", "01"),
				Postcode:    "F1 1FF",
			},
		}).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", ctx, &uid.CreateCaseRequestBody{
			Type: "hw",
			Donor: uid.DonorDetails{
				Name:     "John Smith",
				Dob:      date.New("2000", "01", "01"),
				Postcode: "F1 1FF",
			},
		}).
		Return(uid.CreateCaseResponse{UID: "M-1111"}, nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK:        "LPA#5",
			SK:        "#DONOR#an-id",
			ID:        "5",
			UID:       "M-1111",
			UpdatedAt: now,
			Donor: actor.Donor{
				FirstNames:  "John",
				LastName:    "Smith",
				DateOfBirth: date.New("2000", "01", "01"),
				Address: place.Address{
					Postcode: "F1 1FF",
				},
			},
			Type:                           page.LpaTypeHealthWelfare,
			HasSentApplicationUpdatedEvent: true,
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, uidClient: uidClient, eventClient: eventClient, now: func() time.Time { return now }}

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
	assert.Nil(t, err)
}

func TestDonorStorePutWhenUIDFails(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", ctx, mock.Anything).
		Return(uid.CreateCaseResponse{UID: "M-1111"}, expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, uidClient: uidClient, logger: logger, now: time.Now}

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
	assert.Nil(t, err)
}

func TestDonorStorePutWhenPreviousApplicationLinked(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "previous-application-linked", previousApplicationLinkedEvent{
			UID:                       "M-1111",
			ApplicationReason:         "remake",
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
			ApplicationReason:                     page.RemakeOfInvalidApplication,
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
		ApplicationReason:              page.RemakeOfInvalidApplication,
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
		ApplicationReason:                     page.RemakeOfInvalidApplication,
		PreviousApplicationNumber:             "5555",
		HasSentApplicationUpdatedEvent:        true,
		HasSentPreviousApplicationLinkedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenPreviousApplicationLinkedWhenError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "previous-application-linked", mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, logger: logger, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		ApplicationReason:              page.RemakeOfInvalidApplication,
		PreviousApplicationNumber:      "5555",
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenEvidenceFormRequired(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK:        "LPA#5",
			SK:        "#DONOR#an-id",
			ID:        "5",
			UID:       "M-1111",
			UpdatedAt: now,
			Donor: actor.Donor{
				FirstNames: "John",
				LastName:   "Smithe",
			},
			EvidenceFormAddress: place.Address{
				Line1:      "line1",
				Line2:      "line2",
				Line3:      "line3",
				TownOrCity: "town",
				Postcode:   "post",
			},
			HasSentApplicationUpdatedEvent:   true,
			HasSentEvidenceFormRequiredEvent: true,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "evidence-form-required", evidenceFormRequiredEvent{
			UID:        "M-1111",
			FirstNames: "John",
			LastName:   "Smithe",
			Address: address{
				Line1:      "line1",
				Line2:      "line2",
				Line3:      "line3",
				TownOrCity: "town",
				Postcode:   "post",
			},
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:  "LPA#5",
		SK:  "#DONOR#an-id",
		ID:  "5",
		UID: "M-1111",
		Donor: actor.Donor{
			FirstNames: "John",
			LastName:   "Smithe",
		},
		EvidenceFormAddress: place.Address{
			Line1:      "line1",
			Line2:      "line2",
			Line3:      "line3",
			TownOrCity: "town",
			Postcode:   "post",
		},
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenEvidenceFormRequiredWontResend(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                               "LPA#5",
		SK:                               "#DONOR#an-id",
		ID:                               "5",
		UID:                              "M-1111",
		EvidenceFormAddress:              place.Address{Line1: "line"},
		HasSentApplicationUpdatedEvent:   true,
		HasSentEvidenceFormRequiredEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenEvidenceFormRequiredWhenError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "evidence-form-required", mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, logger: logger, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		EvidenceFormAddress:            place.Address{Line1: "line"},
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenReducedFeeRequested(t *testing.T) {
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
			FeeType:                        page.HalfFee,
			EvidenceKeys:                   []page.Evidence{{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf", Sent: now}},
			Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
			HasSentApplicationUpdatedEvent: true,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "reduced-fee-requested", reducedFeeRequestedEvent{
			UID:         "M-1111",
			RequestType: "HalfFee",
			Evidence:    []string{"lpa-uid-evidence-a-uid"},
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		FeeType:                        page.HalfFee,
		EvidenceKeys:                   []page.Evidence{{Key: "lpa-uid-evidence-a-uid", Filename: "whatever.pdf"}},
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenReducedFeeRequestedSentAndUnsentFees(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK:        "LPA#5",
			SK:        "#DONOR#an-id",
			ID:        "5",
			UID:       "M-1111",
			UpdatedAt: now,
			FeeType:   page.HalfFee,
			EvidenceKeys: []page.Evidence{
				{Key: "lpa-uid-evidence-a-uid-1", Filename: "whatever.pdf", Sent: now},
				{Key: "lpa-uid-evidence-a-uid-2", Filename: "whenever.pdf", Sent: now},
				{Key: "lpa-uid-evidence-a-uid-3", Filename: "whoever.pdf", Sent: now},
			},
			Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
			HasSentApplicationUpdatedEvent: true,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "reduced-fee-requested", reducedFeeRequestedEvent{
			UID:         "M-1111",
			RequestType: "HalfFee",
			Evidence:    []string{"lpa-uid-evidence-a-uid-1", "lpa-uid-evidence-a-uid-3"},
		}).
		Return(nil)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:      "LPA#5",
		SK:      "#DONOR#an-id",
		ID:      "5",
		UID:     "M-1111",
		FeeType: page.HalfFee,
		EvidenceKeys: []page.Evidence{
			{Key: "lpa-uid-evidence-a-uid-1", Filename: "whatever.pdf"},
			{Key: "lpa-uid-evidence-a-uid-2", Filename: "whenever.pdf", Sent: now},
			{Key: "lpa-uid-evidence-a-uid-3", Filename: "whoever.pdf"},
		},
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		HasSentApplicationUpdatedEvent: true,
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

	donorStore := &donorStore{dynamoClient: dynamoClient, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		EvidenceKeys:                   []page.Evidence{{Key: "lpa-uid-evidence-a-uid-1", Filename: "whatever.pdf", Sent: now}},
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStorePutWhenReducedFeeRequestedWhenError(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &page.Lpa{
			PK:                             "LPA#5",
			SK:                             "#DONOR#an-id",
			ID:                             "5",
			UID:                            "M-1111",
			Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
			EvidenceKeys:                   []page.Evidence{{Sent: now}, {}},
			UpdatedAt:                      now,
			HasSentApplicationUpdatedEvent: true,
		}).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.
		On("Send", ctx, "reduced-fee-requested", mock.Anything).
		Return(expectedError)

	logger := newMockLogger(t)
	logger.
		On("Print", expectedError)

	donorStore := &donorStore{dynamoClient: dynamoClient, eventClient: eventClient, logger: logger, now: func() time.Time { return now }}

	err := donorStore.Put(ctx, &page.Lpa{
		PK:                             "LPA#5",
		SK:                             "#DONOR#an-id",
		ID:                             "5",
		UID:                            "M-1111",
		Tasks:                          page.Tasks{PayForLpa: actor.PaymentTaskPending},
		EvidenceKeys:                   []page.Evidence{{Sent: now}, {}},
		HasSentApplicationUpdatedEvent: true,
	})
	assert.Nil(t, err)
}

func TestDonorStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	now := time.Now()
	lpa := &page.Lpa{PK: "LPA#10100000", SK: "#DONOR#an-id", ID: "10100000", CreatedAt: now, UpdatedAt: now}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", ctx, lpa).
		Return(nil)
	dynamoClient.
		On("Create", ctx, lpaLink{PK: "LPA#10100000", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor}).
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
