package app

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type certificateProviderStore struct {
	dataStore DataStore
	randomInt func(int) int
}

func (s *certificateProviderStore) Create(ctx context.Context, lpa *page.Lpa, donorSessionID string) (*actor.CertificateProvider, error) {
	cp := &actor.CertificateProvider{
		ID: "10" + strconv.Itoa(s.randomInt(100000)),
	}

	lpaPk, err := attributevalue.Marshal("DONOR#" + donorSessionID)
	lpaSk, err := attributevalue.Marshal("#LPA#" + lpa.ID)
	if err != nil {
		return nil, err
	}

	lpa.CertificateProviderID = cp.ID
	cp.LpaID = lpa.ID

	lpaData, err := attributevalue.Marshal(lpa)
	if err != nil {
		return nil, err
	}

	update := &types.Update{
		Key:              map[string]types.AttributeValue{"PK": lpaPk, "SK": lpaSk},
		UpdateExpression: aws.String("SET #Data = :lpa"),
		ExpressionAttributeNames: map[string]string{
			"#Data": "Data",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lpa": lpaData,
		},
	}

	err = s.dataStore.PutTransact(ctx, "CERTIFICATE_PROVIDER#"+cp.ID, "#LPA#"+lpa.ID, cp, update)

	return cp, err
}

func (s *certificateProviderStore) Get(ctx context.Context) (*actor.CertificateProvider, error) {
	data := page.SessionDataFromContext(ctx)

	if data.LpaID == "" || data.CertificateProviderID == "" {
		return nil, errors.New("certificateProviderStore.Get requires LpaID and CertificateProviderId to retrieve")
	}

	var certificateProvider actor.CertificateProvider

	pk := "CERTIFICATE_PROVIDER#" + data.CertificateProviderID
	sk := "#LPA#" + data.LpaID
	if err := s.dataStore.Get(ctx, pk, sk, &certificateProvider); err != nil {
		return nil, err
	}

	return &certificateProvider, nil
}

func (s *certificateProviderStore) Put(ctx context.Context, certificateProvider *actor.CertificateProvider) error {
	certificateProvider.UpdatedAt = time.Now()

	pk := "CERTIFICATE_PROVIDER#" + certificateProvider.ID
	sk := "#LPA#" + certificateProvider.LpaID
	return s.dataStore.Put(ctx, pk, sk, certificateProvider)
}
