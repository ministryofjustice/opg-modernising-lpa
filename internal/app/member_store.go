package app

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type memberStore struct {
	dynamoClient DynamoClient
	uuidString   func() string
	now          func() time.Time
}

func (s *memberStore) CreateMemberInvite(ctx context.Context, organisation *actor.Organisation, firstNames, lastname, email, referenceNumber string, permission actor.Permission) error {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID == "" {
		return errors.New("memberStore.Get requires OrganisationID")
	}

	invite := &actor.MemberInvite{
		PK:               dynamo.OrganisationKey(data.OrganisationID),
		SK:               dynamo.MemberInviteKey(email),
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

func (s *memberStore) DeleteMemberInvite(ctx context.Context, organisationID, email string) error {
	if err := s.dynamoClient.DeleteOne(ctx, dynamo.OrganisationKey(organisationID), dynamo.MemberInviteKey(email)); err != nil {
		return fmt.Errorf("error deleting member invite: %w", err)
	}

	return nil
}

func (s *memberStore) Create(ctx context.Context, firstNames, lastName string) (*actor.Member, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("memberStore.Create requires SessionID")
	}

	if data.Email == "" {
		return nil, errors.New("memberStore.Create requires Email")
	}

	organisationID := s.uuidString()

	member := &actor.Member{
		PK:             dynamo.OrganisationKey(organisationID),
		SK:             dynamo.MemberKey(data.SessionID),
		ID:             s.uuidString(),
		OrganisationID: organisationID,
		Email:          data.Email,
		FirstNames:     firstNames,
		LastName:       lastName,
		CreatedAt:      s.now(),
		UpdatedAt:      s.now(),
		Permission:     actor.PermissionAdmin,
		Status:         actor.StatusActive,
		LastLoggedInAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("error creating member: %w", err)
	}

	link := &organisationLink{
		PK:       member.PK,
		SK:       dynamo.MemberIDKey(member.ID),
		MemberSK: member.SK,
	}

	if err := s.dynamoClient.Create(ctx, link); err != nil {
		return nil, fmt.Errorf("error creating organisation link: %w", err)
	}

	return member, nil
}

func (s *memberStore) CreateFromInvite(ctx context.Context, invite *actor.MemberInvite) error {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("memberStore.CreateFromInvite requires SessionID")
	}

	member := &actor.Member{
		PK:             dynamo.OrganisationKey(invite.OrganisationID),
		SK:             dynamo.MemberKey(data.SessionID),
		CreatedAt:      s.now(),
		UpdatedAt:      s.now(),
		ID:             s.uuidString(),
		OrganisationID: invite.OrganisationID,
		Email:          invite.Email,
		FirstNames:     invite.FirstNames,
		LastName:       invite.LastName,
		Permission:     invite.Permission,
		LastLoggedInAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return fmt.Errorf("error creating organisation member: %w", err)
	}

	if err := s.DeleteMemberInvite(ctx, invite.OrganisationID, invite.Email); err != nil {
		return err
	}

	link := &organisationLink{
		PK:       member.PK,
		SK:       dynamo.MemberIDKey(member.ID),
		MemberSK: member.SK,
	}

	if err := s.dynamoClient.Create(ctx, link); err != nil {
		return fmt.Errorf("error creating organisation link: %w", err)
	}

	return nil
}

func (s *memberStore) InvitedMember(ctx context.Context) (*actor.MemberInvite, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.Email == "" {
		return nil, errors.New("memberStore.InvitedMember requires Email")
	}

	var invitedMember *actor.MemberInvite
	if err := s.dynamoClient.OneBySK(ctx, dynamo.MemberInviteKey(data.Email), &invitedMember); err != nil {
		return nil, err
	}

	return invitedMember, nil
}

func (s *memberStore) InvitedMembers(ctx context.Context) ([]*actor.MemberInvite, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("memberStore.InvitedMembers requires OrganisationID")
	}

	var invitedMembers []*actor.MemberInvite
	if err := s.dynamoClient.AllByPartialSK(ctx, dynamo.OrganisationKey(data.OrganisationID), dynamo.MemberInviteKey(""), &invitedMembers); err != nil {
		return nil, err
	}

	return invitedMembers, nil
}

func (s *memberStore) InvitedMembersByEmail(ctx context.Context) ([]*actor.MemberInvite, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.Email == "" {
		return nil, errors.New("memberStore.InvitedMembersByEmail requires Email")
	}

	var invitedMembers []*actor.MemberInvite
	if err := s.dynamoClient.AllBySK(ctx, dynamo.MemberInviteKey(data.Email), &invitedMembers); err != nil {
		return nil, err
	}

	return invitedMembers, nil
}

func (s *memberStore) GetAll(ctx context.Context) ([]*actor.Member, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("memberStore.GetAll requires OrganisationID")
	}

	var members []*actor.Member
	if err := s.dynamoClient.AllByPartialSK(ctx, dynamo.OrganisationKey(data.OrganisationID), dynamo.MemberKey(""), &members); err != nil {
		return nil, err
	}

	slices.SortFunc(members, func(a, b *actor.Member) int {
		return strings.Compare(a.FirstNames, b.FirstNames)
	})

	return members, nil
}

func (s *memberStore) GetByID(ctx context.Context, memberID string) (*actor.Member, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("memberStore.GetByID requires OrganisationID")
	}

	var link *organisationLink
	if err := s.dynamoClient.One(ctx, dynamo.OrganisationKey(data.OrganisationID), dynamo.MemberIDKey(memberID), &link); err != nil {
		return nil, err
	}

	var member *actor.Member
	if err := s.dynamoClient.One(ctx, link.PK, link.MemberSK, &member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *memberStore) Get(ctx context.Context) (*actor.Member, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("memberStore.Get requires SessionID")
	}

	if data.OrganisationID == "" {
		return nil, errors.New("memberStore.Get requires OrganisationID")
	}

	var member *actor.Member
	if err := s.dynamoClient.One(ctx, dynamo.OrganisationKey(data.OrganisationID), dynamo.MemberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *memberStore) GetAny(ctx context.Context) (*actor.Member, error) {
	data, err := appcontext.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("memberStore.Get requires SessionID")
	}

	var member *actor.Member
	if err := s.dynamoClient.OneBySK(ctx, dynamo.MemberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	return member, nil
}

func (s *memberStore) Put(ctx context.Context, member *actor.Member) error {
	member.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, member)
}
