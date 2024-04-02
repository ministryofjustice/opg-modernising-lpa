package lpastore

import (
	"context"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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
	lpa.Tasks = donor.Tasks

	return lpa, nil
}
