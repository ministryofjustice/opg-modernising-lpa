package app

import (
	"context"
	"errors"
	"log"
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

func (s *certificateProviderStore) Create(ctx context.Context, lpa *page.Lpa) (*actor.CertificateProvider, error) {
	cp := &actor.CertificateProvider{
		ID: "10" + strconv.Itoa(s.randomInt(100000)),
	}

	lpaPkSk, err := attributevalue.Marshal("LPA#" + lpa.ID)
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
		Key:              map[string]types.AttributeValue{"PK": lpaPkSk, "SK": lpaPkSk},
		UpdateExpression: aws.String("SET #Data = :lpa"),
		ExpressionAttributeNames: map[string]string{
			"#Data": "Data",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":lpa": lpaData,
		},
	}

	log.Println("Creating CP. CP ID is: ", cp.ID)
	log.Println("Creating CP. LPA ID is: ", lpa.ID)

	err = s.dataStore.PutTransact(ctx, "CERTIFICATE_PROVIDER#"+cp.ID, "LPA#"+lpa.ID, cp, update)

	return cp, err
}

func (s *certificateProviderStore) Get(ctx context.Context) (*actor.CertificateProvider, error) {
	data := page.SessionDataFromContext(ctx)
	if data.LpaID == "" || data.ActorID == "" {
		return nil, errors.New("certificateProviderStore.Get requires LpaID and ActorID to retrieve")
	}

	var certificateProvider actor.CertificateProvider

	log.Println("Getting CP, LPA ID is: " + data.LpaID)
	log.Println("Getting CP, Actor ID is: " + data.ActorID)

	pk := "CERTIFICATE_PROVIDER#" + data.ActorID
	sk := "LPA#" + data.LpaID
	if err := s.dataStore.Get(ctx, pk, sk, &certificateProvider); err != nil {
		return nil, err
	}

	return &certificateProvider, nil
}

func (s *certificateProviderStore) Put(ctx context.Context, certificateProvider *actor.CertificateProvider) error {
	certificateProvider.UpdatedAt = time.Now()

	log.Println("Putting CP, LPA ID is: " + certificateProvider.LpaID)

	pk := "CERTIFICATE_PROVIDER#" + certificateProvider.ID
	sk := "LPA#" + certificateProvider.LpaID
	return s.dataStore.Put(ctx, pk, sk, certificateProvider)
}
