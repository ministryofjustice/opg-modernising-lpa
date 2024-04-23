package lpastore

import (
	"context"
	"errors"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DonorStore interface {
	GetAny(ctx context.Context) (*actor.DonorProvidedDetails, error)
}

type LpaClient interface {
	Lpa(ctx context.Context, lpaUID string) (*Lpa, error)
	Lpas(ctx context.Context, lpaUIDs []string) ([]*Lpa, error)
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

	if donor.LpaUID == "" {
		return FromDonorProvidedDetails(donor), nil
	}

	lpa, err := s.client.Lpa(ctx, donor.LpaUID)
	if errors.Is(err, ErrNotFound) {
		return FromDonorProvidedDetails(donor), nil
	} else if err != nil {
		return nil, err
	}

	return s.merge(lpa, donor), nil
}

func (s *ResolvingService) ResolveList(ctx context.Context, donors []*actor.DonorProvidedDetails) ([]*Lpa, error) {
	lpaUIDs := make([]string, len(donors))
	for i, donor := range donors {
		lpaUIDs[i] = donor.LpaUID
	}

	lpas, err := s.client.Lpas(ctx, lpaUIDs)
	if err != nil {
		return nil, err
	}

	lpaMap := map[string]*Lpa{}
	for _, lpa := range lpas {
		lpaMap[lpa.LpaUID] = lpa
	}

	result := make([]*Lpa, len(donors))
	for i, donor := range donors {
		if lpa, ok := lpaMap[donor.LpaUID]; ok {
			result[i] = s.merge(lpa, donor)
		} else {
			result[i] = FromDonorProvidedDetails(donor)
		}
	}

	return result, nil
}

func (s *ResolvingService) merge(lpa *Lpa, donor *actor.DonorProvidedDetails) *Lpa {
	lpa.LpaID = donor.LpaID
	lpa.LpaUID = donor.LpaUID
	if donor.SK == dynamo.DonorKey("PAPER") {
		lpa.Submitted = true
		lpa.Paid = true
		// set to Professionally so we always show the certificate provider home
		// address question
		lpa.CertificateProvider.Relationship = actor.Professionally
		lpa.Donor.Channel = actor.ChannelPaper
	} else {
		lpa.DonorIdentityConfirmed = donor.DonorIdentityConfirmed()
		lpa.Submitted = !donor.SubmittedAt.IsZero()
		lpa.Paid = donor.Tasks.PayForLpa.IsCompleted()
		lpa.IsOrganisationDonor = strings.HasPrefix(donor.SK, dynamo.OrganisationKey(""))
		lpa.Donor.Channel = actor.ChannelOnline

		// copy the relationship as it isn't stored in the lpastore.
		lpa.CertificateProvider.Relationship = donor.CertificateProvider.Relationship

		if lpa.WithdrawnAt.IsZero() {
			lpa.WithdrawnAt = donor.WithdrawnAt
		}
	}

	return lpa
}
