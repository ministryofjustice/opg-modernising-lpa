package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type dashboardStore struct {
	dataStore DataStore
}

func (s *dashboardStore) GetAll(ctx context.Context) (donor, attorney, certificateProvider []*page.Lpa, err error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	if data.SessionID == "" {
		return nil, nil, nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var keys []struct {
		PK   string
		SK   string
		Data string
	}
	if err := s.dataStore.GetAllByGsi(ctx, "ActorIndex", "#SUB#"+data.SessionID, &keys); err != nil {
		return nil, nil, nil, err
	}

	searchKeys := make([]dynamo.Key, len(keys))
	keyMap := map[string]string{}
	for i, key := range keys {
		sk, actorType, _ := strings.Cut(key.Data, "|")
		searchKeys[i] = dynamo.Key{PK: key.PK, SK: sk}
		keyMap[key.PK] = actorType
	}

	result, err := s.dataStore.GetAllByKeys(ctx, searchKeys)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("what: %w", err)
	}

	for _, item := range result {
		var v struct {
			PK   string
			Data *page.Lpa
		}

		if err := attributevalue.UnmarshalMap(item, &v); err != nil {
			return nil, nil, nil, fmt.Errorf("hey %w", err)
		}

		switch keyMap[v.PK] {
		case "DONOR":
			donor = append(donor, v.Data)
		case "ATTORNEY":
			attorney = append(attorney, v.Data)
		case "CERTIFICATE_PROVIDER":
			certificateProvider = append(certificateProvider, v.Data)
		}
	}

	return donor, attorney, certificateProvider, nil
}
