package app

import (
	"context"
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type evidenceReceivedStore struct {
	dynamoClient DynamoClient
}

func (s *evidenceReceivedStore) Get(ctx context.Context) (bool, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return false, err
	}

	if data.LpaID == "" {
		return false, errors.New("evidenceReceivedStore.Get requires LpaID")
	}

	var v any
	if err := s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), dynamo.EvidenceReceivedKey(), &v); err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
