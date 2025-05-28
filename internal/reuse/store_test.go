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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("hi")

func (c *mockDynamoClient_One_Call) SetData(data map[string]types.AttributeValue) {
	c.Run(func(_ context.Context, _ dynamo.PK, _ dynamo.SK, v any) {
		attributevalue.UnmarshalMap(data, v)
	})
}

func TestStoreDeleteReusable(t *testing.T) {
	actorUID := actoruid.New()

	testcases := map[actor.Type]func(s *Store, ctx context.Context) error{
		actor.TypeCorrespondent: func(s *Store, ctx context.Context) error {
			return s.DeleteCorrespondent(ctx, donordata.Correspondent{UID: actorUID})
		},
		actor.TypeAttorney: func(s *Store, ctx context.Context) error {
			return s.DeleteAttorney(ctx, donordata.Attorney{UID: actorUID})
		},
		actor.TypeTrustCorporation: func(s *Store, ctx context.Context) error {
			return s.DeleteTrustCorporation(ctx, donordata.TrustCorporation{UID: actorUID})
		},
		actor.TypeCertificateProvider: func(s *Store, ctx context.Context) error {
			return s.DeleteCertificateProvider(ctx, donordata.CertificateProvider{UID: actorUID})
		},
	}

	for actorType, fn := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Update(ctx, dynamo.ReuseKey("session-id", actorType.String()), dynamo.MetadataKey(""),
					map[string]string{"#ActorUID": actorUID.String()},
					map[string]types.AttributeValue(nil),
					"REMOVE #ActorUID",
				).
				Return(expectedError)

			err := fn(NewStore(dynamoClient), ctx)
			assert.Equal(t, expectedError, err)
		})

		t.Run(actorType.String()+"/MissingSession", func(t *testing.T) {
			ctx := context.Background()

			err := fn(NewStore(nil), ctx)
			assert.Equal(t, appcontext.SessionMissingError{}, err)
		})

		t.Run(actorType.String()+"/MissingSessionID", func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

			err := fn(NewStore(nil), ctx)
			assert.Error(t, err)
		})
	}
}

func TestStorePutReusable(t *testing.T) {
	actorUID := actoruid.New()

	testcases := map[actor.Type]struct {
		fn             func(s *Store, ctx context.Context, v any) error
		item           any
		withoutAddress any
		updated        any
	}{
		actor.TypeCorrespondent: {
			fn: func(s *Store, ctx context.Context, v any) error {
				return s.PutCorrespondent(ctx, v.(donordata.Correspondent))
			},
			item:    donordata.Correspondent{UID: actorUID, FirstNames: "John"},
			updated: donordata.Correspondent{FirstNames: "John"},
		},
		actor.TypeAttorney: {
			fn: func(s *Store, ctx context.Context, v any) error {
				return s.PutAttorney(ctx, v.(donordata.Attorney))
			},
			item:           donordata.Attorney{UID: actorUID, FirstNames: "John", Address: place.Address{Line1: "a"}},
			withoutAddress: donordata.Attorney{UID: actorUID, FirstNames: "John"},
			updated:        donordata.Attorney{FirstNames: "John", Address: place.Address{Line1: "a"}},
		},
		actor.TypeTrustCorporation: {
			fn: func(s *Store, ctx context.Context, v any) error {
				return s.PutTrustCorporation(ctx, v.(donordata.TrustCorporation))
			},
			item:           donordata.TrustCorporation{UID: actorUID, Name: "Corp", Address: place.Address{Line1: "a"}},
			withoutAddress: donordata.TrustCorporation{UID: actorUID, Name: "Corp"},
			updated:        donordata.TrustCorporation{Name: "Corp", Address: place.Address{Line1: "a"}},
		},
		actor.TypeCertificateProvider: {
			fn: func(s *Store, ctx context.Context, v any) error {
				return s.PutCertificateProvider(ctx, v.(donordata.CertificateProvider))
			},
			item:           donordata.CertificateProvider{UID: actorUID, FirstNames: "John", Address: place.Address{Line1: "a"}},
			withoutAddress: donordata.CertificateProvider{UID: actorUID, FirstNames: "John"},
			updated:        donordata.CertificateProvider{FirstNames: "John", Address: place.Address{Line1: "a"}},
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})
			value, _ := attributevalue.Marshal(tc.updated)

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Update(ctx, dynamo.ReuseKey("session-id", actorType.String()), dynamo.MetadataKey(""),
					map[string]string{"#ActorUID": actorUID.String()},
					map[string]types.AttributeValue{":Value": value},
					"SET #ActorUID = :Value",
				).
				Return(expectedError)

			err := tc.fn(NewStore(dynamoClient), ctx, tc.item)
			assert.Equal(t, expectedError, err)
		})

		t.Run(actorType.String()+"/Supporter", func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id", OrganisationID: "org"})

			err := tc.fn(NewStore(nil), ctx, tc.item)
			assert.Nil(t, err)
		})

		t.Run(actorType.String()+"/MissingSession", func(t *testing.T) {
			ctx := context.Background()

			err := tc.fn(NewStore(nil), ctx, tc.item)
			assert.Equal(t, appcontext.SessionMissingError{}, err)
		})

		t.Run(actorType.String()+"/MissingSessionID", func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

			err := tc.fn(NewStore(nil), ctx, tc.item)
			assert.Error(t, err)
		})

		if tc.withoutAddress != nil {
			t.Run(actorType.String()+"/MissingAddress", func(t *testing.T) {
				ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

				err := tc.fn(NewStore(nil), ctx, tc.withoutAddress)
				assert.Nil(t, err)
			})
		}
	}
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

func TestStoreCertificateProviders(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	expected := []donordata.CertificateProvider{
		{FirstNames: "Adam"},
		{FirstNames: "Dave"},
		{FirstNames: "John"},
	}

	marshalled0, _ := attributevalue.Marshal(expected[0])
	marshalled1, _ := attributevalue.Marshal(expected[1])
	marshalled2, _ := attributevalue.Marshal(expected[2])

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(ctx, dynamo.ReuseKey("session-id", actor.TypeCertificateProvider.String()), dynamo.MetadataKey(""), mock.Anything).
		Return(nil).
		SetData(map[string]types.AttributeValue{
			"PK":    &types.AttributeValueMemberS{Value: "REUSE#session-id"},
			"SK":    &types.AttributeValueMemberS{Value: "METADATA#"},
			"uid-a": marshalled2,
			"uid-b": marshalled0,
			"uid-c": marshalled1,
			"uid-d": marshalled0,
		})

	result, err := NewStore(dynamoClient).CertificateProviders(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expected, result)
}

func TestStoreCertificateProvidersWhenDynamoErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	_, err := NewStore(dynamoClient).CertificateProviders(ctx)
	assert.Equal(t, expectedError, err)
}

func TestStoreCertificateProvidersWhenMissingSession(t *testing.T) {
	ctx := context.Background()

	_, err := NewStore(nil).CertificateProviders(ctx)
	assert.Equal(t, appcontext.SessionMissingError{}, err)
}

func TestStoreCertificateProvidersWhenMissingSessionID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{})

	_, err := NewStore(nil).CertificateProviders(ctx)
	assert.Error(t, err)
}
