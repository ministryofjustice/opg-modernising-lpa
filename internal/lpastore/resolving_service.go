package lpastore

import (
	"context"
	"errors"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DonorStore interface {
	GetAny(ctx context.Context) (*actor.DonorProvidedDetails, error)
}

type LpaClient interface {
	Lpa(ctx context.Context, lpaUID string) (*Lpa, error)
}

// A ResolvingService wraps a Client so that an Lpa can be retrieved without
// passing its UID.
type ResolvingService struct {
	donorStore DonorStore
	client     LpaClient
}

func NewResolvingService(donorStore DonorStore, client LpaClient) *ResolvingService {
	return &ResolvingService{donorStore: donorStore, client: client}
}

func (s *ResolvingService) Get(ctx context.Context) (*Lpa, error) {
	donor, err := s.donorStore.GetAny(ctx)
	if err != nil {
		return nil, err
	}

	lpa, err := s.client.Lpa(ctx, donor.LpaUID)
	if errors.Is(err, ErrNotFound) {
		lpa = &Lpa{}
	} else if err != nil {
		return nil, err
	}

	lpa.LpaID = donor.LpaID
	lpa.LpaUID = donor.LpaUID
	if donor.SK == dynamo.DonorKey("PAPER") {
		lpa.DonorIdentityConfirmed = true
		lpa.Submitted = true
		lpa.Paid = true
		// set to Professionally so we always show the certificate provider home
		// address question
		lpa.CertificateProvider.Relationship = actor.Professionally
	} else {
		lpa.DonorIdentityConfirmed = donor.DonorIdentityConfirmed()
		lpa.Submitted = !donor.SubmittedAt.IsZero()
		lpa.Paid = donor.Tasks.PayForLpa.IsCompleted()
		lpa.IsOrganisationDonor = strings.HasPrefix(donor.SK, dynamo.OrganisationKey(""))
		lpa.CertificateProvider.Relationship = donor.CertificateProvider.Relationship
		// TODO: eventually we'll need to remove the RegisteredAt field as mlpa
		// won't be tracking that data, then we'll need to figure out how to expose
		// the data for testing
		if !donor.RegisteredAt.IsZero() {
			lpa.RegisteredAt = date.FromTime(donor.RegisteredAt)
		}
	}

	return lpa, nil
}
