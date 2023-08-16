package app

import (
	"context"
	"errors"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type evidenceReceivedStore struct {
	dynamoClient DynamoClient
}

func (s *evidenceReceivedStore) Get(ctx context.Context) (bool, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return false, err
	}

	if data.LpaID == "" {
		return false, errors.New("evidenceReceivedStore.Get requires LpaID")
	}

	var v any
	if err := s.dynamoClient.Get(ctx, lpaKey(data.LpaID), "#EVIDENCE_RECEIVED", &v); err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
