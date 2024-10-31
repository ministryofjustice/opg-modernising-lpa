package dynamo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testPK string

func (k testPK) PK() string { return string(k) }

type testSK string

func (k testSK) SK() string { return string(k) }

var (
	ctx           = context.Background()
	expectedError = errors.New("err")
)

func TestOne(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{Item: data}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var actual map[string]string
	err := c.One(ctx, testPK("a-pk"), testSK("a-sk"), &actual)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestOneWhenError(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.One(ctx, testPK("a-pk"), testSK("a-sk"), &v)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", v)
}

func TestOneWhenNotFound(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.One(ctx, testPK("a-pk"), testSK("a-sk"), &v)
	assert.Equal(t, NotFoundError{}, err)
	assert.Equal(t, "", v)
}

func TestOneByUID(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(lpaUIDIndex),
			ExpressionAttributeNames:  map[string]string{"#LpaUID": "LpaUID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#LpaUID = :LpaUID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{{
				"PK":     &types.AttributeValueMemberS{Value: "LPA#123"},
				"LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
			}},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]any
	err := c.OneByUID(ctx, "M-1111-2222-3333", &v)

	assert.Nil(t, err)
	assert.Equal(t, map[string]any{"PK": "LPA#123", "LpaUID": "M-1111-2222-3333"}, v)
}

func TestOneByUIDWhenQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(lpaUIDIndex),
			ExpressionAttributeNames:  map[string]string{"#LpaUID": "LpaUID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#LpaUID = :LpaUID"),
		}).
		Return(&dynamodb.QueryOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.OneByUID(ctx, "M-1111-2222-3333", mock.Anything)

	assert.Equal(t, fmt.Errorf("failed to query UID: %w", expectedError), err)
}

func TestOneByUIDWhenNot1Item(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(lpaUIDIndex),
			ExpressionAttributeNames:  map[string]string{"#LpaUID": "LpaUID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#LpaUID = :LpaUID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK":     &types.AttributeValueMemberS{Value: "LPA#123"},
					"LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
				},
				{
					"PK":     &types.AttributeValueMemberS{Value: "LPA#123"},
					"LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
				},
			},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.OneByUID(ctx, "M-1111-2222-3333", mock.Anything)

	assert.Equal(t, errors.New("expected to resolve LpaUID but got 2 items"), err)
}

func TestOneByUIDWhenUnmarshalError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(lpaUIDIndex),
			ExpressionAttributeNames:  map[string]string{"#LpaUID": "LpaUID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#LpaUID = :LpaUID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK":     &types.AttributeValueMemberS{Value: "LPA#123"},
					"LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
				},
			},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.OneByUID(ctx, "M-1111-2222-3333", "not an lpa")

	assert.IsType(t, &attributevalue.InvalidUnmarshalError{}, err)
}

func TestOneByPK(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	pkey, _ := attributevalue.Marshal("a-pk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey},
			KeyConditionExpression:    aws.String("#PK = :PK"),
			Limit:                     aws.Int32(1),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPK(ctx, testPK("a-pk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestOneByPKOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPK(ctx, testPK("a-pk"), &v)
	assert.Equal(t, expectedError, err)
}

func TestOneByPKWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPK(ctx, testPK("a-pk"), &v)
	assert.Equal(t, NotFoundError{}, err)
}

func TestOneByPartialSK(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
			Limit:                     aws.Int32(1),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSK(ctx, testPK("a-pk"), testSK("a-partial-sk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestOneByPartialSKOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSK(ctx, testPK("a-pk"), testSK("a-partial-sk"), &v)
	assert.Equal(t, expectedError, err)
}

func TestOneByPartialSKWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSK(ctx, testPK("a-pk"), testSK("a-partial-sk"), &v)
	assert.Equal(t, NotFoundError{}, err)
}

func TestAllByPartialSK(t *testing.T) {
	ctx := context.Background()

	expected := []map[string]string{{"Col": "Val"}, {"Other": "Thing"}}
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected[0])
	data2, _ := attributevalue.MarshalMap(expected[1])

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data2}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []map[string]string
	err := c.AllByPartialSK(ctx, testPK("a-pk"), testSK("a-partial-sk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestAllByPartialSKOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.AllByPartialSK(ctx, testPK("a-pk"), testSK("a-partial-sk"), &v)
	assert.Equal(t, expectedError, err)
}

func TestAllForActor(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(skUpdatedAtIndex),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey},
			KeyConditionExpression:    aws.String("#SK = :SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []map[string]string
	err := c.AllBySK(ctx, testSK("a-partial-sk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, []map[string]string{expected, expected}, v)
}

func TestAllForActorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.AllBySK(ctx, testSK("a-partial-sk"), &v)
	assert.Nil(t, err)
	assert.Empty(t, v)
}

func TestAllForActorOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.AllBySK(ctx, testSK("a-partial-sk"), &v)
	assert.Equal(t, expectedError, err)
}

func TestLatestForActor(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	skey, _ := attributevalue.Marshal("a-partial-sk")
	updated, _ := attributevalue.Marshal("2")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(skUpdatedAtIndex),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK", "#UpdatedAt": "UpdatedAt"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey, ":UpdatedAt": updated},
			KeyConditionExpression:    aws.String("#SK = :SK and #UpdatedAt > :UpdatedAt"),
			ScanIndexForward:          aws.Bool(false),
			Limit:                     aws.Int32(1),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.LatestForActor(ctx, testSK("a-partial-sk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestLatestForActorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v interface{}
	err := c.LatestForActor(ctx, testSK("a-partial-sk"), &v)
	assert.Nil(t, err)
	assert.Nil(t, v)
}

func TestLatestForActorOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.LatestForActor(ctx, testSK("a-partial-sk"), &v)
	assert.Equal(t, expectedError, err)
}

func TestAllKeysByPK(t *testing.T) {
	ctx := context.Background()

	keys := []Keys{
		{PK: LpaKey("pk"), SK: OrganisationKey("sk1")},
		{PK: LpaKey("pk"), SK: DonorKey("sk2")},
	}

	item1, _ := attributevalue.MarshalMap(keys[0])
	item2, _ := attributevalue.MarshalMap(keys[1])

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                aws.String("this"),
			ExpressionAttributeNames: map[string]string{"#PK": "PK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":PK": &types.AttributeValueMemberS{Value: "LPA#pk"},
			},
			KeyConditionExpression: aws.String("#PK = :PK"),
			ProjectionExpression:   aws.String("PK, SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item1, item2}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	result, err := c.AllKeysByPK(ctx, LpaKey("pk"))
	assert.Nil(t, err)
	assert.Equal(t, keys, result)
}

func TestAllKeysByPKWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	_, err := c.AllKeysByPK(ctx, testPK("pk"))
	assert.Equal(t, expectedError, err)
}

func TestAllByKeys(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				"this": {
					Keys: []map[string]types.AttributeValue{{
						"PK": &types.AttributeValueMemberS{Value: "pk"},
						"SK": &types.AttributeValueMemberS{Value: "sk"},
					}},
				},
			},
		}).
		Return(&dynamodb.BatchGetItemOutput{
			Responses: map[string][]map[string]types.AttributeValue{
				"this": {data},
			},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	v, err := c.AllByKeys(ctx, []Keys{{PK: testPK("pk"), SK: testSK("sk")}})
	assert.Nil(t, err)
	assert.Equal(t, []map[string]types.AttributeValue{data}, v)
}

func TestAllByKeysWhenQueryErrors(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		BatchGetItem(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	_, err := c.AllByKeys(ctx, []Keys{{PK: testPK("pk"), SK: testSK("sk")}})
	assert.Equal(t, expectedError, err)
}

func TestPut(t *testing.T) {
	testCases := map[string]map[string]string{
		"Without UpdatedAt": {"Col": "Val"},
		"Zero UpdatedAt":    {"Col": "Val", "UpdatedAt": "0001-01-01T00:00:00Z"},
	}

	for name, dataMap := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data, _ := attributevalue.MarshalMap(dataMap)

			dynamoDB := newMockDynamoDB(t)
			dynamoDB.EXPECT().
				PutItem(ctx, &dynamodb.PutItemInput{
					TableName: aws.String("this"),
					Item:      data,
				}).
				Return(&dynamodb.PutItemOutput{}, nil)

			c := &Client{table: "this", svc: dynamoDB}

			err := c.Put(ctx, dataMap)
			assert.Nil(t, err)
		})
	}
}

func TestPutWhenStructHasVersion(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]any{"Col": "Val", "Version": 2})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, &dynamodb.PutItemInput{
			TableName:                 aws.String("this"),
			Item:                      data,
			ConditionExpression:       aws.String("Version = :version"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":version": &types.AttributeValueMemberN{Value: "1"}},
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, map[string]any{"Col": "Val", "Version": 1})
	assert.Nil(t, err)
}

func TestPutWhenConditionalCheckFailedException(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]any{"Col": "Val", "Version": 2})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, &dynamodb.PutItemInput{
			TableName:                 aws.String("this"),
			Item:                      data,
			ConditionExpression:       aws.String("Version = :version"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":version": &types.AttributeValueMemberN{Value: "1"}},
		}).
		Return(&dynamodb.PutItemOutput{}, &smithy.OperationError{Err: &types.ConditionalCheckFailedException{}})

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, map[string]any{"Col": "Val", "Version": 1})
	assert.Equal(t, ConditionalCheckFailedError{}, err)
}

func TestPutWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, mock.Anything).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, "hello")
	assert.Equal(t, expectedError, err)
}

func TestPutWhenUnmarshalError(t *testing.T) {
	ctx := context.Background()

	c := &Client{table: "this", svc: newMockDynamoDB(t)}

	err := c.Put(ctx, map[string]string{"Col": "Val", "Version": "not an int"})
	assert.NotNil(t, err)
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]string{"Col": "Val"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, &dynamodb.PutItemInput{
			TableName:           aws.String("this"),
			Item:                data,
			ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Create(ctx, map[string]string{"Col": "Val"})
	assert.Nil(t, err)
}

func TestCreateWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, mock.Anything).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Create(ctx, map[string]string{"Col": "Val"})
	assert.Equal(t, expectedError, err)
}

func TestCreateOnly(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]string{"Col": "Val"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, &dynamodb.PutItemInput{
			TableName:           aws.String("this"),
			Item:                data,
			ConditionExpression: aws.String("attribute_not_exists(PK)"),
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.CreateOnly(ctx, map[string]string{"Col": "Val"})
	assert.Nil(t, err)
}

func TestCreateOnlyWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		PutItem(ctx, mock.Anything).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.CreateOnly(ctx, map[string]string{"Col": "Val"})
	assert.Equal(t, expectedError, err)
}

func TestDeleteKeys(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Delete: &types.Delete{
						TableName: aws.String("this"),
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "pk"},
							"SK": &types.AttributeValueMemberS{Value: "sk1"},
						},
					},
				},
				{
					Delete: &types.Delete{
						TableName: aws.String("this"),
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "pk"},
							"SK": &types.AttributeValueMemberS{Value: "sk2"},
						},
					},
				},
			},
		}).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.DeleteKeys(ctx, []Keys{{PK: testPK("pk"), SK: testSK("sk1")}, {PK: testPK("pk"), SK: testSK("sk2")}})
	assert.Equal(t, expectedError, err)
}

func TestDeleteOne(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String("table-name"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "a-pk"},
				"SK": &types.AttributeValueMemberS{Value: "a-sk"},
			},
		}).
		Return(nil, expectedError)

	c := &Client{table: "table-name", svc: dynamoDB}

	err := c.DeleteOne(ctx, testPK("a-pk"), testSK("a-sk"))

	assert.Equal(t, expectedError, err)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String("table-name"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "a-pk"},
				"SK": &types.AttributeValueMemberS{Value: "a-sk"},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{"prop": &types.AttributeValueMemberS{Value: "val"}},
			UpdateExpression:          aws.String("some = expression"),
		}).
		Return(nil, nil)

	c := &Client{table: "table-name", svc: dynamoDB}

	err := c.Update(ctx, testPK("a-pk"), testSK("a-sk"), map[string]types.AttributeValue{"prop": &types.AttributeValueMemberS{Value: "val"}}, "some = expression")

	assert.Nil(t, err)
}

func TestUpdateOnServiceError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String("table-name"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "a-pk"},
				"SK": &types.AttributeValueMemberS{Value: "a-sk"},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{"Col": &types.AttributeValueMemberS{Value: "Val"}},
			UpdateExpression:          aws.String("some = expression"),
		}).
		Return(nil, expectedError)

	c := &Client{table: "table-name", svc: dynamoDB}

	err := c.Update(ctx, testPK("a-pk"), testSK("a-sk"), map[string]types.AttributeValue{"Col": &types.AttributeValueMemberS{Value: "Val"}}, "some = expression")

	assert.Equal(t, expectedError, err)
}

func TestUpdateReturn(t *testing.T) {
	ctx := context.Background()

	returned, _ := attributevalue.MarshalMap(map[string]any{"a": "b"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String("table-name"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "a-pk"},
				"SK": &types.AttributeValueMemberS{Value: "a-sk"},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{"prop": &types.AttributeValueMemberS{Value: "val"}},
			UpdateExpression:          aws.String("some = expression"),
			ReturnValues:              types.ReturnValueAllNew,
		}).
		Return(&dynamodb.UpdateItemOutput{Attributes: returned}, nil)

	c := &Client{table: "table-name", svc: dynamoDB}

	result, err := c.UpdateReturn(ctx, testPK("a-pk"), testSK("a-sk"), map[string]types.AttributeValue{"prop": &types.AttributeValueMemberS{Value: "val"}}, "some = expression")
	assert.Nil(t, err)
	assert.Equal(t, returned, result)
}

func TestUpdateReturnOnServiceError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		UpdateItem(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "table-name", svc: dynamoDB}

	_, err := c.UpdateReturn(ctx, testPK("a-pk"), testSK("a-sk"), map[string]types.AttributeValue{"Col": &types.AttributeValueMemberS{Value: "Val"}}, "some = expression")

	assert.Equal(t, expectedError, err)
}

func TestBatchPutOneBatch(t *testing.T) {
	ctx := context.Background()

	values := []any{map[string]string{"a": "b"}, map[string]string{"x": "y"}}
	itemA, _ := attributevalue.MarshalMap(values[0])
	itemB, _ := attributevalue.MarshalMap(values[1])

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Put: &types.Put{
						TableName: aws.String("table-name"),
						Item:      itemA,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String("table-name"),
						Item:      itemB,
					},
				},
			},
		}).
		Return(nil, expectedError)

	c := &Client{table: "table-name", svc: dynamoDB}
	err := c.BatchPut(ctx, values)

	assert.Equal(t, expectedError, err)
}

func TestOneBySk(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	skey, _ := attributevalue.Marshal("sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(skUpdatedAtIndex),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey},
			KeyConditionExpression:    aws.String("#SK = :SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneBySK(ctx, testSK("sk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestOneBySKWhenNotOneResult(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	data, _ := attributevalue.MarshalMap(expected)

	testcases := map[string]struct {
		items         []map[string]types.AttributeValue
		expectedError error
	}{
		"no results": {
			expectedError: NotFoundError{},
		},
		"multiple results": {
			items:         []map[string]types.AttributeValue{data, data},
			expectedError: errors.New("expected to resolve SK but got 2 items"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoDB := newMockDynamoDB(t)
			dynamoDB.EXPECT().
				Query(mock.Anything, mock.Anything).
				Return(&dynamodb.QueryOutput{Items: tc.items}, nil)

			c := &Client{table: "this", svc: dynamoDB}

			var v map[string]string
			err := c.OneBySK(ctx, testSK("sk"), &v)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestOneBySkWhenQueryError(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(mock.Anything, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneBySK(ctx, testSK("sk"), &v)

	assert.Equal(t, expectedError, err)
}

func TestMove(t *testing.T) {
	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Delete: &types.Delete{
						TableName: aws.String("this"),
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "a-pk"},
							"SK": &types.AttributeValueMemberS{Value: "an-sk"},
						},
						ConditionExpression: aws.String("attribute_exists(PK) and attribute_exists(SK)"),
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String("this"),
						Item: map[string]types.AttributeValue{
							"hey": &types.AttributeValueMemberS{Value: "hi"},
						},
					},
				},
			},
		}).
		Return(nil, nil)

	c := &Client{table: "this", svc: dynamoDB}
	err := c.Move(ctx, Keys{PK: testPK("a-pk"), SK: testSK("an-sk")}, map[string]string{"hey": "hi"})
	assert.Nil(t, err)
}

func TestMoveWhenConflict(t *testing.T) {
	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(mock.Anything, mock.Anything).
		Return(nil, &types.TransactionConflictException{})

	c := &Client{table: "this", svc: dynamoDB}
	err := c.Move(ctx, Keys{PK: testPK("a-pk"), SK: testSK("an-sk")}, map[string]string{"hey": "hi"})
	assert.Equal(t, ConditionalCheckFailedError{}, err)
}

func TestMoveWhenConditionalCheckFailed(t *testing.T) {
	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(mock.Anything, mock.Anything).
		Return(nil, &types.TransactionCanceledException{
			CancellationReasons: []types.CancellationReason{
				{Code: aws.String("ConditionalCheckFailed")},
			},
		})

	c := &Client{table: "this", svc: dynamoDB}
	err := c.Move(ctx, Keys{PK: testPK("a-pk"), SK: testSK("an-sk")}, map[string]string{"hey": "hi"})
	assert.Equal(t, ConditionalCheckFailedError{}, err)
}

func TestMoveWhenOtherCancellation(t *testing.T) {
	canceledException := &types.TransactionCanceledException{
		CancellationReasons: []types.CancellationReason{
			{Code: aws.String("What")},
		},
	}

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(mock.Anything, mock.Anything).
		Return(nil, canceledException)

	c := &Client{table: "this", svc: dynamoDB}
	err := c.Move(ctx, Keys{PK: testPK("a-pk"), SK: testSK("an-sk")}, map[string]string{"hey": "hi"})
	assert.Equal(t, canceledException, err)
}

func TestAnyByPK(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	pkey, _ := attributevalue.Marshal("a-pk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey},
			KeyConditionExpression:    aws.String("#PK = :PK"),
			Limit:                     aws.Int32(1),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.AnyByPK(ctx, testPK("a-pk"), &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestAnyByPKOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.AnyByPK(ctx, testPK("a-pk"), &v)
	assert.Equal(t, expectedError, err)
}

func TestAnyByPKWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		Query(ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.AnyByPK(ctx, testPK("a-pk"), &v)
	assert.Equal(t, NotFoundError{}, err)
}
