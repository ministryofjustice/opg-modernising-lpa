package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type organisationStore struct {
	dynamoClient DynamoClient
	uuidString   func() string
	randomString func(int) string
	now          func() time.Time
}

func (s *organisationStore) Create(ctx context.Context, name string) (*actor.Organisation, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Create requires SessionID")
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
		PK:        memberKey(data.SessionID),
		SK:        organisationKey(organisationID),
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("error creating organisation member: %w", err)
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
	if err := s.dynamoClient.OneByPartialSK(ctx, memberKey(data.SessionID), organisationKey(""), &member); err != nil {
		return nil, err
	}

	var organisation actor.Organisation
	if err := s.dynamoClient.One(ctx, member.SK, member.SK, &organisation); err != nil {
		return nil, err
	}

	return &organisation, nil
}

func (s *organisationStore) Put(ctx context.Context, organisation *actor.Organisation) error {
	organisation.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, organisation)
}

func (s *organisationStore) CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, code string, permission actor.Permission) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID == "" {
		return errors.New("organisationStore.Get requires OrganisationID")
	}

	invite := &actor.MemberInvite{
		PK:             organisationKey(data.OrganisationID),
		SK:             memberInviteKey(code),
		CreatedAt:      s.now(),
		OrganisationID: organisation.ID,
		Email:          email,
		FirstNames:     firstNames,
		LastName:       lastname,
		Permission:     permission,
	}

	if err := s.dynamoClient.Create(ctx, invite); err != nil {
		return fmt.Errorf("error creating member invite: %w", err)
	}

	return nil
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

func (s *organisationStore) InvitedMembers(ctx context.Context) ([]*actor.MemberInvite, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("organisationStore.Get requires OrganisationID")
	}

	var invitedMembers []*actor.MemberInvite
	if err := s.dynamoClient.AllByPartialSk(ctx, organisationKey(data.OrganisationID), memberInviteKey(""), &invitedMembers); err != nil {
		return nil, err
	}

	return invitedMembers, nil
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
	if err := s.dynamoClient.AllForActor(ctx, organisationKey(data.OrganisationID), &donors); err != nil {
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

func organisationKey(s string) string {
	return "ORGANISATION#" + s
}

func memberKey(s string) string {
	return "MEMBER#" + s
}

func memberInviteKey(s string) string {
	return "MEMBERINVITE#" + s
}
