package dynamo

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

type testItem struct {
	PK    UIDKeyType
	SK    MetadataKeyType
	Value int
}

var (
	item1 = testItem{PK: UIDKey("some-pk"), SK: MetadataKey("some-sk"), Value: 1}
	item2 = testItem{PK: UIDKey("some-pk2"), SK: MetadataKey("some-sk2"), Value: 2}
	item3 = testItem{PK: UIDKey("some-pk"), SK: MetadataKey("some-sk3"), Value: 3}
	item4 = testItem{PK: UIDKey("some-pk3"), SK: MetadataKey("some-sk"), Value: 4}
)

func withClient(t *testing.T, fn func(client *Client)) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("eu-west-1"),
		config.WithBaseEndpoint("http://localhost:4566"),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "fakeKeyId",
				SecretAccessKey: "fakeAccessKey",
			}, nil
		})),
	)
	if err != nil {
		t.Logf("unable to load SDK config: %s", err)
		t.Fail()
		return
	}

	client, _ := NewClient(cfg, "test")
	dynamo := client.svc.(*dynamodb.Client)
	_, err = dynamo.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("test"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	})
	if !assert.Nil(t, err) {
		return
	}

	defer dynamo.DeleteTable(ctx, &dynamodb.DeleteTableInput{TableName: aws.String("test")})

	fn(client)
}

func createTestItems(client *Client) error {
	transaction := NewTransaction()

	for _, item := range []testItem{item1, item2, item3, item4} {
		transaction.Put(item)
	}

	return client.WriteTransaction(ctx, transaction)
}

func TestIntegrationClientOne(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	withClient(t, func(client *Client) {
		assert.Nil(t, createTestItems(client))

		var v testItem
		err := client.One(ctx, UIDKey("some-pk"), MetadataKey("some-sk"), &v)
		assert.Nil(t, err)
		assert.Equal(t, 1, v.Value)
	})
}

func TestIntegrationClientAllKeysByPK(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	withClient(t, func(client *Client) {
		assert.Nil(t, createTestItems(client))

		keys, err := client.AllKeysByPK(ctx, UIDKey("some-pk"))
		assert.Nil(t, err)
		assert.Equal(t, []Keys{
			{PK: UIDKey("some-pk"), SK: MetadataKey("some-sk")},
			{PK: UIDKey("some-pk"), SK: MetadataKey("some-sk3")},
		}, keys)
	})
}

func TestIntegrationClientAllByKeys(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	withClient(t, func(client *Client) {
		assert.Nil(t, createTestItems(client))

		values, err := client.AllByKeys(ctx, []Keys{
			{PK: UIDKey("some-pk"), SK: MetadataKey("some-sk")},
			{PK: UIDKey("some-pk"), SK: MetadataKey("some-sk3")},
		})
		assert.Nil(t, err)

		var v []testItem
		_ = attributevalue.UnmarshalListOfMaps(values, &v)
		assert.Equal(t, []testItem{item3, item1}, v)
	})
}

func TestIntegrationClientOneByPK(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	withClient(t, func(client *Client) {
		assert.Nil(t, createTestItems(client))

		var v testItem
		err := client.OneByPK(ctx, UIDKey("some-pk"), &v)
		assert.Nil(t, err)
		assert.Equal(t, item1, v)
	})
}

func TestIntegrationClientCreate(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	withClient(t, func(client *Client) {
		assert.Nil(t, createTestItems(client))

		t.Run("create existing (PK,SK)", func(t *testing.T) {
			err := client.Create(ctx, item1)
			v := &types.ConditionalCheckFailedException{}
			assert.ErrorAs(t, err, &v)
		})

		t.Run("create existing PK", func(t *testing.T) {
			err := client.Create(ctx, testItem{PK: item1.PK, SK: MetadataKey("other")})
			assert.Nil(t, err)
		})

		t.Run("create existing SK", func(t *testing.T) {
			err := client.Create(ctx, testItem{PK: UIDKey("other"), SK: item1.SK})
			assert.Nil(t, err)
		})
	})
}

func TestIntegrationClientTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	withClient(t, func(client *Client) {
		assert.Nil(t, createTestItems(client))

		t.Run("create existing (PK,SK)", func(t *testing.T) {
			transaction := NewTransaction().Create(item1)
			err := client.WriteTransaction(ctx, transaction)
			assert.ErrorIs(t, err, ConditionalCheckFailedError{})
		})

		t.Run("create existing PK", func(t *testing.T) {
			transaction := NewTransaction().Create(testItem{PK: item1.PK, SK: MetadataKey("other")})
			err := client.WriteTransaction(ctx, transaction)
			assert.Nil(t, err)
		})

		t.Run("create existing SK", func(t *testing.T) {
			transaction := NewTransaction().Create(testItem{PK: UIDKey("other"), SK: item1.SK})
			err := client.WriteTransaction(ctx, transaction)
			assert.Nil(t, err)
		})
	})
}
