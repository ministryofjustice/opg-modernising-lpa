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
