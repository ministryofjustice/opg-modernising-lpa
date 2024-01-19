package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type organisationStore struct {
	dynamoClient DynamoClient
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

	nameHash := sha256.Sum256([]byte(name))
	organisationID := hex.EncodeToString(nameHash[:])

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
		PK:        organisationKey(organisationID),
		SK:        subKey(data.SessionID),
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return fmt.Errorf("error creating organisation member: %w", err)
	}

	return nil
}

func organisationKey(s string) string {
	return "ORGANISATION#" + s
}
