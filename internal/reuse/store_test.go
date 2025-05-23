package reuse

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("hi")

func (c *mockDynamoClient_One_Call) SetData(data map[string]types.AttributeValue) {
	c.Run(func(_ context.Context, _ dynamo.PK, _ dynamo.SK, v any) {
		attributevalue.UnmarshalMap(data, v)
	})
}

func TestStorePutCorrespondent(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	actorUID := actoruid.New()
	value, _ := attributevalue.Marshal(donordata.Correspondent{FirstNames: "John"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(ctx, dynamo.ReuseKey("session-id", actor.TypeCorrespondent.String()), dynamo.MetadataKey(""),
			map[string]string{"#ActorUID": actorUID.String()},
			map[string]types.AttributeValue{":Value": value},
			"SET #ActorUID = :Value",
		).
		Return(expectedError)

	err := NewStore(dynamoClient).PutCorrespondent(ctx, donordata.Correspondent{UID: actorUID, FirstNames: "John"})
	assert.Equal(t, expectedError, err)
}

func TestStorePutCorrespondentWhenSupporter(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id", OrganisationID: "org"})

	err := NewStore(nil).PutCorrespondent(ctx, donordata.Correspondent{})
	assert.Nil(t, err)
}

func TestStorePutCorrespondentWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	err := NewStore(nil).PutCorrespondent(ctx, donordata.Correspondent{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStorePutCorrespondentWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	err := NewStore(nil).PutCorrespondent(ctx, donordata.Correspondent{})
	assert.Error(t, err)
}

func TestStoreDeleteCorrespondent(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	actorUID := actoruid.New()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(ctx, dynamo.ReuseKey("session-id", actor.TypeCorrespondent.String()), dynamo.MetadataKey(""),
			map[string]string{"#ActorUID": actorUID.String()},
			map[string]types.AttributeValue(nil),
			"REMOVE #ActorUID",
		).
		Return(expectedError)

	err := NewStore(dynamoClient).DeleteCorrespondent(ctx, donordata.Correspondent{UID: actorUID, FirstNames: "John"})
	assert.Equal(t, expectedError, err)
}

func TestStoreDeleteCorrespondentWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	err := NewStore(nil).DeleteCorrespondent(ctx, donordata.Correspondent{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreDeleteCorrespondentWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	err := NewStore(nil).DeleteCorrespondent(ctx, donordata.Correspondent{})
	assert.Error(t, err)
}

func TestStoreCorrespondents(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	expected := []donordata.Correspondent{
		{FirstNames: "Adam"},
		{FirstNames: "Dave"},
		{FirstNames: "John"},
	}

	marshalled0, _ := attributevalue.Marshal(expected[0])
	marshalled1, _ := attributevalue.Marshal(expected[1])
	marshalled2, _ := attributevalue.Marshal(expected[2])

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(ctx, dynamo.ReuseKey("session-id", actor.TypeCorrespondent.String()), dynamo.MetadataKey(""), mock.Anything).
		Return(nil).
		SetData(map[string]types.AttributeValue{
			"PK":    &types.AttributeValueMemberS{Value: "REUSE#session-id"},
			"SK":    &types.AttributeValueMemberS{Value: "METADATA#"},
			"uid-a": marshalled2,
			"uid-b": marshalled0,
			"uid-c": marshalled1,
			"uid-d": marshalled0,
		})

	result, err := NewStore(dynamoClient).Correspondents(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestStoreCorrespondentsWhenDynamoErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	_, err := NewStore(dynamoClient).Correspondents(ctx)
	assert.Equal(t, expectedError, err)
}

func TestStoreCorrespondentsWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	_, err := NewStore(nil).Correspondents(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreCorrespondentsWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	_, err := NewStore(nil).Correspondents(ctx)
	assert.Error(t, err)
}

func TestStorePutAttorneys(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	actorUID0, actorUID1 := actoruid.New(), actoruid.New()
	value0, _ := attributevalue.Marshal(donordata.Attorney{FirstNames: "John"})
	value1, _ := attributevalue.Marshal(donordata.Attorney{FirstNames: "Barry"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(ctx, dynamo.ReuseKey("session-id", actor.TypeAttorney.String()), dynamo.MetadataKey(""),
			map[string]string{"#ActorUID0": actorUID0.String(), "#ActorUID1": actorUID1.String()},
			map[string]types.AttributeValue{":Value0": value0, ":Value1": value1},
			"SET #ActorUID0 = :Value0, #ActorUID1 = :Value1",
		).
		Return(expectedError)

	err := NewStore(dynamoClient).PutAttorneys(ctx, []donordata.Attorney{
		{UID: actorUID0, FirstNames: "John"},
		{UID: actorUID1, FirstNames: "Barry"},
	})
	assert.Equal(t, expectedError, err)
}

func TestStorePutAttorneysWhenSupporter(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id", OrganisationID: "org"})

	err := NewStore(nil).PutAttorneys(ctx, []donordata.Attorney{})
	assert.Nil(t, err)
}

func TestStorePutAttorneysWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	err := NewStore(nil).PutAttorneys(ctx, []donordata.Attorney{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStorePutAttorneysWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	err := NewStore(nil).PutAttorneys(ctx, []donordata.Attorney{})
	assert.Error(t, err)
}

func TestStoreDeleteAttorney(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	actorUID := actoruid.New()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(ctx, dynamo.ReuseKey("session-id", actor.TypeAttorney.String()), dynamo.MetadataKey(""),
			map[string]string{"#ActorUID": actorUID.String()},
			map[string]types.AttributeValue(nil),
			"REMOVE #ActorUID",
		).
		Return(expectedError)

	err := NewStore(dynamoClient).DeleteAttorney(ctx, donordata.Attorney{UID: actorUID, FirstNames: "John"})
	assert.Equal(t, expectedError, err)
}

func TestStoreDeleteAttorneyWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	err := NewStore(nil).DeleteAttorney(ctx, donordata.Attorney{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreDeleteAttorneyWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	err := NewStore(nil).DeleteAttorney(ctx, donordata.Attorney{})
	assert.Error(t, err)
}

func TestStoreAttorneys(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	existingAttorney := donordata.Attorney{FirstNames: "Barry"}
	existingReplacementAttorney := donordata.Attorney{FirstNames: "Charles"}

	expected := []donordata.Attorney{
		{FirstNames: "Adam"},
		{FirstNames: "Dave"},
		{FirstNames: "John"},
	}

	marshalled0, _ := attributevalue.Marshal(expected[0])
	marshalled1, _ := attributevalue.Marshal(expected[1])
	marshalled2, _ := attributevalue.Marshal(expected[2])
	marshalled3, _ := attributevalue.Marshal(existingAttorney)
	marshalled4, _ := attributevalue.Marshal(existingReplacementAttorney)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(ctx, dynamo.ReuseKey("session-id", actor.TypeAttorney.String()), dynamo.MetadataKey(""), mock.Anything).
		Return(nil).
		SetData(map[string]types.AttributeValue{
			"PK":    &types.AttributeValueMemberS{Value: "REUSE#session-id"},
			"SK":    &types.AttributeValueMemberS{Value: "METADATA#"},
			"uid-a": marshalled2,
			"uid-b": marshalled0,
			"uid-c": marshalled1,
			"uid-d": marshalled0,
			"uid-e": marshalled3,
			"uid-f": marshalled4,
		})

	result, err := NewStore(dynamoClient).Attorneys(ctx, &donordata.Provided{
		Attorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{existingAttorney},
		},
		ReplacementAttorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{existingReplacementAttorney},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestStoreAttorneysWhenDynamoErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	_, err := NewStore(dynamoClient).Attorneys(ctx, &donordata.Provided{})
	assert.Equal(t, expectedError, err)
}

func TestStoreAttorneysWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	_, err := NewStore(nil).Attorneys(ctx, &donordata.Provided{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreAttorneysWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	_, err := NewStore(nil).Attorneys(ctx, &donordata.Provided{})
	assert.Error(t, err)
}

func TestStorePutTrustCorporation(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	actorUID := actoruid.New()
	value, _ := attributevalue.Marshal(donordata.TrustCorporation{Name: "Corp"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(ctx, dynamo.ReuseKey("session-id", actor.TypeTrustCorporation.String()), dynamo.MetadataKey(""),
			map[string]string{"#ActorUID": actorUID.String()},
			map[string]types.AttributeValue{":Value": value},
			"SET #ActorUID = :Value",
		).
		Return(expectedError)

	err := NewStore(dynamoClient).PutTrustCorporation(ctx, donordata.TrustCorporation{UID: actorUID, Name: "Corp"})
	assert.Equal(t, expectedError, err)
}

func TestStorePutTrustCorporationWhenSupporter(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id", OrganisationID: "org"})

	err := NewStore(nil).PutTrustCorporation(ctx, donordata.TrustCorporation{})
	assert.Nil(t, err)
}

func TestStorePutTrustCorporationWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	err := NewStore(nil).PutTrustCorporation(ctx, donordata.TrustCorporation{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStorePutTrustCorporationWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	err := NewStore(nil).PutTrustCorporation(ctx, donordata.TrustCorporation{})
	assert.Error(t, err)
}

func TestStoreDeleteTrustCorporation(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
	actorUID := actoruid.New()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Update(ctx, dynamo.ReuseKey("session-id", actor.TypeTrustCorporation.String()), dynamo.MetadataKey(""),
			map[string]string{"#ActorUID": actorUID.String()},
			map[string]types.AttributeValue(nil),
			"REMOVE #ActorUID",
		).
		Return(expectedError)

	err := NewStore(dynamoClient).DeleteTrustCorporation(ctx, donordata.TrustCorporation{UID: actorUID, Name: "Corp"})
	assert.Equal(t, expectedError, err)
}

func TestStoreDeleteTrustCorporationWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	err := NewStore(nil).DeleteTrustCorporation(ctx, donordata.TrustCorporation{})
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreDeleteTrustCorporationWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	err := NewStore(nil).DeleteTrustCorporation(ctx, donordata.TrustCorporation{})
	assert.Error(t, err)
}

func TestStoreTrustCorporations(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	expected := []donordata.TrustCorporation{
		{Name: "Corp"},
		{Name: "Trust"},
		{Name: "Untrustworthy"},
	}

	marshalled0, _ := attributevalue.Marshal(expected[0])
	marshalled1, _ := attributevalue.Marshal(expected[1])
	marshalled2, _ := attributevalue.Marshal(expected[2])

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(ctx, dynamo.ReuseKey("session-id", actor.TypeTrustCorporation.String()), dynamo.MetadataKey(""), mock.Anything).
		Return(nil).
		SetData(map[string]types.AttributeValue{
			"PK":    &types.AttributeValueMemberS{Value: "REUSE#session-id"},
			"SK":    &types.AttributeValueMemberS{Value: "METADATA#"},
			"uid-a": marshalled2,
			"uid-b": marshalled0,
			"uid-c": marshalled1,
			"uid-d": marshalled0,
		})

	result, err := NewStore(dynamoClient).TrustCorporations(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestStoreTrustCorporationsWhenDynamoErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	_, err := NewStore(dynamoClient).TrustCorporations(ctx)
	assert.Equal(t, expectedError, err)
}

func TestStoreTrustCorporationsWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	_, err := NewStore(nil).TrustCorporations(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreTrustCorporationsWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	_, err := NewStore(nil).TrustCorporations(ctx)
	assert.Error(t, err)
}
