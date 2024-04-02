package lpastore

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DonorStore interface {
	GetAny(ctx context.Context) (*actor.DonorProvidedDetails, error)
}

// A ResolvingService wraps a Client so that an Lpa can be retrieved without
// passing its UID.
type ResolvingService struct {
	donorStore DonorStore
	client     *Client
}

func NewResolvingService(donorStore DonorStore, client *Client) *ResolvingService {
	return &ResolvingService{donorStore: donorStore, client: client}
}

func (s *ResolvingService) Get(ctx context.Context) (*actor.DonorProvidedDetails, error) {
	donor, err := s.donorStore.GetAny(ctx)
	if err != nil {
		return nil, err
	}

	lpa, err := s.client.Lpa(ctx, donor.LpaUID)
	if err != nil {
		return nil, err
	}

	lpa.LpaID = donor.LpaID
	lpa.LpaUID = donor.LpaUID
	if donor.SK == dynamo.DonorKey("PAPER") {
		// set these tasks completed as they are completed, and are used for
		// certificate provider logic
		lpa.Tasks = actor.DonorTasks{
			PayForLpa:                  actor.PaymentTaskCompleted,
			ConfirmYourIdentityAndSign: actor.TaskCompleted,
		}

		// set to Professionally so we always show the certificate provider home
		// address question
		lpa.CertificateProvider.Relationship = actor.Professionally
	} else {
		lpa.Tasks = donor.Tasks
		lpa.CertificateProvider.Relationship = donor.CertificateProvider.Relationship
	}

	return lpa, nil
}
