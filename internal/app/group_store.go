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

type groupStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *groupStore) Create(ctx context.Context, name string) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("groupStore.Create requires SessionID")
	}

	nameHash := sha256.Sum256([]byte(name))
	groupID := hex.EncodeToString(nameHash[:])

	group := &actor.Group{
		PK:        groupKey(groupID),
		SK:        groupKey(groupID),
		ID:        groupID,
		Name:      name,
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, group); err != nil {
		return fmt.Errorf("error creating group: %w", err)
	}

	member := &actor.GroupMember{
		PK:        groupKey(groupID),
		SK:        subKey(data.SessionID),
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, member); err != nil {
		return fmt.Errorf("error creating group member: %w", err)
	}

	return nil
}

func groupKey(s string) string {
	return "GROUP#" + s
}
