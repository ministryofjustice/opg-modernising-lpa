package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type organisationStore struct {
	dynamoClient DynamoClient
	uuidString   func() string
	newUID       func() actoruid.UID
	randomString func(int) string
	now          func() time.Time
}

// An organisationLink is used to join a Member to an Organisation to be accessed by MemberID.
type organisationLink struct {
	// PK is the same as the PK for the Member
	PK dynamo.OrganisationKeyType
	// SK is the Member ID for the Member
	SK       dynamo.MemberIDKeyType
	MemberSK dynamo.MemberKeyType
}

func (s *organisationStore) Create(ctx context.Context, member *actor.Member, name string) (*actor.Organisation, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Create requires SessionID")
	}

	organisation := &actor.Organisation{
		PK:        dynamo.OrganisationKey(member.OrganisationID),
		SK:        dynamo.OrganisationKey(member.OrganisationID),
		ID:        member.OrganisationID,
		Name:      name,
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, organisation); err != nil {
		return nil, fmt.Errorf("error creating organisation: %w", err)
	}

	return organisation, nil
}

func (s *organisationStore) Get(ctx context.Context) (*actor.Organisation, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Get requires SessionID")
	}

	var member actor.Member
	if err := s.dynamoClient.OneBySK(ctx, dynamo.MemberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	var organisation actor.Organisation
	if err := s.dynamoClient.One(ctx, member.PK, member.PK, &organisation); err != nil {
		return nil, err
	}

	if !organisation.DeletedAt.IsZero() {
		return nil, dynamo.NotFoundError{}
	}

	return &organisation, err
}

func (s *organisationStore) Put(ctx context.Context, organisation *actor.Organisation) error {
	organisation.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, organisation)
}

func (s *organisationStore) CreateLPA(ctx context.Context) (*actor.DonorProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("organisationStore.CreateLPA requires OrganisationID")
	}

	lpaID := s.uuidString()
	donorUID := s.newUID()

	donor := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey(lpaID),
		SK:        dynamo.LpaOwnerKey(dynamo.OrganisationKey(data.OrganisationID)),
		LpaID:     lpaID,
		CreatedAt: s.now(),
		Version:   1,
		Donor: actor.Donor{
			UID: donorUID,
		},
	}

	if err := donor.UpdateHash(); err != nil {
		return nil, err
	}

	if err := s.dynamoClient.Create(ctx, donor); err != nil {
		return nil, err
	}

	return donor, err
}

func (s *organisationStore) SoftDelete(ctx context.Context, organisation *actor.Organisation) error {
	organisation.DeletedAt = s.now()

	return s.dynamoClient.Put(ctx, organisation)
}
