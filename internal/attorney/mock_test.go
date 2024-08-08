package attorney

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/mock"
)

var (
	ctx           = context.Background()
	expectedError = errors.New("err")
)

func (m *mockDynamoClient) ExpectOne(ctx, pk, sk, data interface{}, err error) {
	m.
		On("One", ctx, pk, sk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (m *mockDynamoClient) ExpectOneByPK(ctx, pk, data interface{}, err error) {
	m.
		On("OneByPK", ctx, pk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		}).
		Once()
}

func (m *mockDynamoClient) ExpectOneByPartialSK(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("OneByPartialSK", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByPartialSK(ctx, pk, partialSk, data interface{}, err error) {
	m.
		On("AllByPartialSK", ctx, pk, partialSk, mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, partialSk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllBySK(ctx, sk, data interface{}, err error) {
	m.
		On("AllBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectLatestForActor(ctx, sk, data interface{}, err error) {
	m.
		On("LatestForActor", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}

func (m *mockDynamoClient) ExpectAllByKeys(ctx context.Context, keys []dynamo.Keys, data []map[string]types.AttributeValue, err error) {
	m.EXPECT().
		AllByKeys(ctx, keys).
		Return(data, err)
}

func (m *mockDynamoClient) ExpectOneBySK(ctx, sk, data interface{}, err error) {
	m.
		On("OneBySK", ctx, sk, mock.Anything).
		Return(func(ctx context.Context, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(data)
			json.Unmarshal(b, v)
			return err
		})
}
