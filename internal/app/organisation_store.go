package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type organisationStore struct {
	dynamoClient DynamoClient
	uuidString   func() string
	now          func() time.Time
}

func (s *organisationStore) Create(ctx context.Context, name string) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("organisationStore.Create requires SessionID")
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
		return fmt.Errorf("error creating organisation: %w", err)
	}

	member := &actor.Member{
		PK:        memberKey(data.SessionID),
		SK:        organisationKey(organisationID),
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return fmt.Errorf("error creating organisation member: %w", err)
	}

	return nil
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
	if err := s.dynamoClient.OneByPartialSk(ctx, memberKey(data.SessionID), organisationKey(""), &member); err != nil {
		return nil, err
	}

	var organisation actor.Organisation
	if err := s.dynamoClient.One(ctx, member.SK, member.SK, &organisation); err != nil {
		return nil, err
	}

	return &organisation, nil
}

func organisationKey(s string) string {
	return "ORGANISATION#" + s
}

func memberKey(s string) string {
	return "MEMBER#" + s
}
