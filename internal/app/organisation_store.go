package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type organisationStore struct {
	dynamoClient DynamoClient
	uuidString   func() string
	randomString func(int) string
	now          func() time.Time
}

// An organisationLink is used to join a Member to an Organisation to be accessed by MemberID.
type organisationLink struct {
	// PK is the same as the PK for the Member
	PK string
	// SK is the Member ID for the Member
	SK       string
	MemberSK string
}

func (s *organisationStore) Create(ctx context.Context, name string) (*actor.Organisation, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Create requires SessionID")
	}

	if data.Email == "" {
		return nil, errors.New("organisationStore.Create requires Email")
	}

	organisationID := s.uuidString()

	organisation := &actor.Organisation{
		PK:        organisationKey(organisationID),
		SK:        organisationKey(organisationID),
		ID:        organisationID,
		Name:      name,
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, organisation); err != nil {
		return nil, fmt.Errorf("error creating organisation: %w", err)
	}

	member := &actor.Member{
		PK:         organisationKey(organisationID),
		SK:         memberKey(data.SessionID),
		ID:         s.uuidString(),
		Email:      data.Email,
		CreatedAt:  s.now(),
		Permission: actor.Admin,
		Status:     actor.Active,
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("error creating organisation member: %w", err)
	}

	link := &organisationLink{
		PK:       member.PK,
		SK:       memberIDKey(member.ID),
		MemberSK: member.SK,
	}

	if err := s.dynamoClient.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("error creating organisation link: %w", err)
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
	if err := s.dynamoClient.OneBySK(ctx, memberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	var organisation actor.Organisation
	if err := s.dynamoClient.One(ctx, member.PK, member.PK, &organisation); err != nil {
		return nil, err
	}

	err = nil
	if !organisation.DeletedAt.IsZero() {
		err = dynamo.NotFoundError{}
		organisation = actor.Organisation{}
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

	donor := &actor.DonorProvidedDetails{
		PK:        lpaKey(lpaID),
		SK:        organisationKey(data.OrganisationID),
		LpaID:     lpaID,
		CreatedAt: s.now(),
		Version:   1,
	}

	if donor.Hash, err = donor.GenerateHash(); err != nil {
		return nil, err
	}

	if err := s.dynamoClient.Create(ctx, donor); err != nil {
		return nil, err
	}

	return donor, err
}

func (s *organisationStore) AllLPAs(ctx context.Context) ([]actor.DonorProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("organisationStore.AllLPAs requires OrganisationID")
	}

	var donors []actor.DonorProvidedDetails
	if err := s.dynamoClient.AllBySK(ctx, organisationKey(data.OrganisationID), &donors); err != nil {
		return nil, fmt.Errorf("organisationStore.AllLPAs error retrieving keys for organisation: %w", err)
	}

	donors = slices.DeleteFunc(donors, func(donor actor.DonorProvidedDetails) bool {
		return !strings.HasPrefix(donor.PK, lpaKey("")) || donor.LpaUID == ""
	})

	slices.SortFunc(donors, func(a, b actor.DonorProvidedDetails) int {
		return strings.Compare(a.Donor.FullName(), b.Donor.FullName())
	})

	return donors, nil
}

func (s *organisationStore) SoftDelete(ctx context.Context) error {
	organisation, err := s.Get(ctx)
	if err != nil {
		return err
	}

	organisation.DeletedAt = s.now()

	return s.dynamoClient.Put(ctx, organisation)
}

func organisationKey(organisationID string) string {
	return "ORGANISATION#" + organisationID
}
