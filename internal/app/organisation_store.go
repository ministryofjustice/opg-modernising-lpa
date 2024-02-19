package app

import (
	"context"
	"encoding/base64"
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
		PK:        organisationKey(organisationID),
		SK:        memberKey(data.SessionID),
		Email:     data.Email,
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
	if err := s.dynamoClient.OneBySK(ctx, memberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	var organisation actor.Organisation
	if err := s.dynamoClient.One(ctx, member.PK, member.PK, &organisation); err != nil {
		return nil, err
	}

	return &organisation, nil
}

func (s *organisationStore) Put(ctx context.Context, organisation *actor.Organisation) error {
	organisation.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, organisation)
}

func (s *organisationStore) PutMember(ctx context.Context, member *actor.Member) error {
	member.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, member)
}

func (s *organisationStore) CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, referenceNumber string, permission actor.Permission) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID == "" {
		return errors.New("organisationStore.Get requires OrganisationID")
	}

	invite := &actor.MemberInvite{
		PK:               organisationKey(data.OrganisationID),
		SK:               memberInviteKey(email),
		CreatedAt:        s.now(),
		OrganisationID:   organisation.ID,
		OrganisationName: organisation.Name,
		Email:            email,
		FirstNames:       firstNames,
		LastName:         lastname,
		Permission:       permission,
		ReferenceNumber:  referenceNumber,
	}

	if err := s.dynamoClient.Create(ctx, invite); err != nil {
		return fmt.Errorf("error creating member invite: %w", err)
	}

	return nil
}

func (s *organisationStore) CreateMember(ctx context.Context, invite *actor.MemberInvite) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("organisationStore.CreateMember requires SessionID")
	}

	member := &actor.Member{
		PK:         organisationKey(invite.OrganisationID),
		SK:         memberKey(data.SessionID),
		CreatedAt:  s.now(),
		UpdatedAt:  s.now(),
		Email:      invite.Email,
		FirstNames: invite.FirstNames,
		LastName:   invite.LastName,
		Permission: invite.Permission,
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return fmt.Errorf("error creating organisation member: %w", err)
	}

	if err := s.dynamoClient.DeleteOne(ctx, invite.PK, invite.SK); err != nil {
		return fmt.Errorf("error deleting member invite: %w", err)
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
		return nil, errors.New("organisationStore.InvitedMembers requires OrganisationID")
	}

	var invitedMembers []*actor.MemberInvite
	if err := s.dynamoClient.AllByPartialSk(ctx, organisationKey(data.OrganisationID), memberInviteKey(""), &invitedMembers); err != nil {
		return nil, err
	}

	return invitedMembers, nil
}

func (s *organisationStore) InvitedMember(ctx context.Context) (*actor.MemberInvite, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.Email == "" {
		return nil, errors.New("organisationStore.InvitedMember requires Email")
	}

	var invitedMember *actor.MemberInvite
	if err := s.dynamoClient.OneBySK(ctx, memberInviteKey(data.Email), &invitedMember); err != nil {
		return nil, err
	}

	return invitedMember, nil
}

func (s *organisationStore) Members(ctx context.Context) ([]*actor.Member, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("organisationStore.Members requires OrganisationID")
	}

	var members []*actor.Member
	if err := s.dynamoClient.AllByPartialSk(ctx, organisationKey(data.OrganisationID), memberKey(""), &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (s *organisationStore) Member(ctx context.Context) (*actor.Member, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Member requires SessionID")
	}

	if data.OrganisationID == "" {
		return nil, errors.New("organisationStore.Member requires OrganisationID")
	}

	var member *actor.Member
	if err := s.dynamoClient.One(ctx, organisationKey(data.OrganisationID), memberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	return member, nil
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

func organisationKey(organisationID string) string {
	return "ORGANISATION#" + organisationID
}

func memberKey(sessionID string) string {
	return "MEMBER#" + sessionID
}

func memberInviteKey(email string) string {
	return fmt.Sprintf("MEMBERINVITE#%s", base64.StdEncoding.EncodeToString([]byte(email)))
}
