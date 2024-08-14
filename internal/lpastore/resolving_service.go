package lpastore

import (
	"context"
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

type DonorStore interface {
	GetAny(ctx context.Context) (*donordata.Provided, error)
}

type LpaClient interface {
	Lpa(ctx context.Context, lpaUID string) (*lpadata.Lpa, error)
	Lpas(ctx context.Context, lpaUIDs []string) ([]*lpadata.Lpa, error)
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

func (s *ResolvingService) Get(ctx context.Context) (*lpadata.Lpa, error) {
	donor, err := s.donorStore.GetAny(ctx)
	if err != nil {
		return nil, err
	}

	if donor.LpaUID == "" {
		return s.merge(FromDonorProvidedDetails(donor), donor), nil
	}

	lpa, err := s.client.Lpa(ctx, donor.LpaUID)
	if errors.Is(err, ErrNotFound) {
		lpa = FromDonorProvidedDetails(donor)
	} else if err != nil {
		return nil, err
	}

	return s.merge(lpa, donor), nil
}

func (s *ResolvingService) ResolveList(ctx context.Context, donors []*donordata.Provided) ([]*lpadata.Lpa, error) {
	lpaUIDs := make([]string, len(donors))
	for i, donor := range donors {
		lpaUIDs[i] = donor.LpaUID
	}

	lpas, err := s.client.Lpas(ctx, lpaUIDs)
	if err != nil {
		return nil, err
	}

	lpaMap := map[string]*lpadata.Lpa{}
	for _, lpa := range lpas {
		lpaMap[lpa.LpaUID] = lpa
	}

	result := make([]*lpadata.Lpa, len(donors))
	for i, donor := range donors {
		if lpa, ok := lpaMap[donor.LpaUID]; ok {
			result[i] = s.merge(lpa, donor)
		} else {
			result[i] = s.merge(FromDonorProvidedDetails(donor), donor)
		}
	}

	return result, nil
}

func (s *ResolvingService) merge(lpa *lpadata.Lpa, donor *donordata.Provided) *lpadata.Lpa {
	lpa.LpaKey = donor.PK
	lpa.LpaOwnerKey = donor.SK
	lpa.LpaID = donor.LpaID
	lpa.LpaUID = donor.LpaUID
	lpa.PerfectAt = donor.PerfectAt
	if donor.SK.Equals(dynamo.DonorKey("PAPER")) {
		lpa.Drafted = true
		lpa.Submitted = true
		lpa.Paid = true
		// set to Professionally so we always show the certificate provider home
		// address question
		lpa.CertificateProvider.Relationship = lpadata.Professionally
		lpa.Donor.Channel = lpadata.ChannelPaper
	} else {
		lpa.Drafted = donor.Tasks.CheckYourLpa.IsCompleted()
		lpa.Submitted = !donor.SubmittedAt.IsZero()
		lpa.Paid = donor.Tasks.PayForLpa.IsCompleted()
		_, lpa.IsOrganisationDonor = donor.SK.Organisation()
		lpa.Donor.Channel = lpadata.ChannelOnline
		lpa.Correspondent = lpadata.Correspondent{
			FirstNames: donor.Correspondent.FirstNames,
			LastName:   donor.Correspondent.LastName,
			Email:      donor.Correspondent.Email,
		}

		// copy the relationship as it isn't stored in the lpastore.
		lpa.CertificateProvider.Relationship = donor.CertificateProvider.Relationship

		if lpa.WithdrawnAt.IsZero() {
			lpa.WithdrawnAt = donor.WithdrawnAt
		}
	}

	return lpa
}
