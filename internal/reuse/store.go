// Package reuse handles storing and retrieving reusable information for LPAs.
package reuse

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v any) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, names map[string]string, values map[string]types.AttributeValue, expression string) error
}

type Store struct {
	dynamoClient DynamoClient
}

func NewStore(dynamoClient DynamoClient) *Store {
	return &Store{
		dynamoClient: dynamoClient,
	}
}

func (s *Store) PutCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error {
	actorUID := correspondent.UID
	correspondent.UID = actoruid.UID{}

	return putReusable(ctx, s.dynamoClient, actor.TypeCorrespondent, actorUID, correspondent)
}

func (s *Store) DeleteCorrespondent(ctx context.Context, correspondent donordata.Correspondent) error {
	return deleteReusable(ctx, s.dynamoClient, actor.TypeCorrespondent, correspondent.UID)
}

func (s *Store) Correspondents(ctx context.Context) ([]donordata.Correspondent, error) {
	correspondents, err := reusables[donordata.Correspondent](ctx, s.dynamoClient, actor.TypeCorrespondent, nil)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(correspondents, func(a, b donordata.Correspondent) int {
		return strings.Compare(a.FullName(), b.FullName())
	})

	return correspondents, nil
}

func (s *Store) PutAttorney(ctx context.Context, attorney donordata.Attorney) error {
	if attorney.Address.Line1 == "" {
		return nil
	}

	actorUID := attorney.UID
	attorney.UID = actoruid.UID{}

	return putReusable(ctx, s.dynamoClient, actor.TypeAttorney, actorUID, attorney)
}

func (s *Store) PutAttorneys(ctx context.Context, attorneys []donordata.Attorney) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID != "" {
		return nil
	}

	if data.SessionID == "" {
		return errors.New("reuseStore.PutAttorneys requires SessionID")
	}

	names := map[string]string{}
	values := map[string]types.AttributeValue{}
	var statements []string
	for i, attorney := range attorneys {
		index := strconv.Itoa(i)

		names["#ActorUID"+index] = attorney.UID.String()

		attorney.UID = actoruid.UID{}
		value, err := attributevalue.Marshal(attorney)
		if err != nil {
			return fmt.Errorf("marshal attorney: %w", err)
		}

		values[":Value"+index] = value

		statements = append(statements, "#ActorUID"+index+" = :Value"+index)
	}

	return s.dynamoClient.Update(ctx, dynamo.ReuseKey(data.SessionID, actor.TypeAttorney.String()), dynamo.MetadataKey(""),
		names,
		values,
		"SET "+strings.Join(statements, ", "),
	)
}

func (s *Store) DeleteAttorney(ctx context.Context, attorney donordata.Attorney) error {
	return deleteReusable(ctx, s.dynamoClient, actor.TypeAttorney, attorney.UID)
}

func (s *Store) Attorneys(ctx context.Context, provided *donordata.Provided) ([]donordata.Attorney, error) {
	seen := map[donordata.Attorney]struct{}{}
	for _, attorney := range provided.Attorneys.Attorneys {
		attorney.UID = actoruid.UID{}
		seen[attorney] = struct{}{}
	}
	for _, attorney := range provided.ReplacementAttorneys.Attorneys {
		attorney.UID = actoruid.UID{}
		seen[attorney] = struct{}{}
	}

	attorneys, err := reusables[donordata.Attorney](ctx, s.dynamoClient, actor.TypeAttorney, seen)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(attorneys, func(a, b donordata.Attorney) int {
		return strings.Compare(a.FullName(), b.FullName())
	})

	return attorneys, nil
}

func (s *Store) PutTrustCorporation(ctx context.Context, trustCorporation donordata.TrustCorporation) error {
	if trustCorporation.Address.Line1 == "" {
		return nil
	}

	actorUID := trustCorporation.UID
	trustCorporation.UID = actoruid.UID{}

	return putReusable(ctx, s.dynamoClient, actor.TypeTrustCorporation, actorUID, trustCorporation)
}

func (s *Store) DeleteTrustCorporation(ctx context.Context, trustCorporation donordata.TrustCorporation) error {
	return deleteReusable(ctx, s.dynamoClient, actor.TypeTrustCorporation, trustCorporation.UID)
}

func (s *Store) TrustCorporations(ctx context.Context, provided *donordata.Provided) ([]donordata.TrustCorporation, error) {
	seen := map[donordata.TrustCorporation]struct{}{}
	if trustCorporation := provided.Attorneys.TrustCorporation; trustCorporation.Name != "" {
		trustCorporation.UID = actoruid.UID{}
		seen[trustCorporation] = struct{}{}
	}
	if trustCorporation := provided.ReplacementAttorneys.TrustCorporation; trustCorporation.Name != "" {
		trustCorporation.UID = actoruid.UID{}
		seen[trustCorporation] = struct{}{}
	}

	trustCorporations, err := reusables[donordata.TrustCorporation](ctx, s.dynamoClient, actor.TypeTrustCorporation, seen)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(trustCorporations, func(a, b donordata.TrustCorporation) int {
		return strings.Compare(a.Name, b.Name)
	})

	return trustCorporations, nil
}

func (s *Store) PutCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error {
	if certificateProvider.Address.Line1 == "" {
		return nil
	}

	actorUID := certificateProvider.UID
	certificateProvider.UID = actoruid.UID{}

	return putReusable(ctx, s.dynamoClient, actor.TypeCertificateProvider, actorUID, certificateProvider)
}

func (s *Store) DeleteCertificateProvider(ctx context.Context, certificateProvider donordata.CertificateProvider) error {
	return deleteReusable(ctx, s.dynamoClient, actor.TypeCertificateProvider, certificateProvider.UID)
}

func (s *Store) CertificateProviders(ctx context.Context) ([]donordata.CertificateProvider, error) {
	certificateProviders, err := reusables[donordata.CertificateProvider](ctx, s.dynamoClient, actor.TypeCertificateProvider, nil)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(certificateProviders, func(a, b donordata.CertificateProvider) int {
		return strings.Compare(a.FullName(), b.FullName())
	})

	return certificateProviders, nil
}

func (s *Store) PutPersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error {
	if personToNotify.Address.Line1 == "" {
		return nil
	}

	actorUID := personToNotify.UID
	personToNotify.UID = actoruid.UID{}

	return putReusable(ctx, s.dynamoClient, actor.TypePersonToNotify, actorUID, personToNotify)
}

func (s *Store) PutPeopleToNotify(ctx context.Context, peopleToNotify []donordata.PersonToNotify) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID != "" {
		return nil
	}

	if data.SessionID == "" {
		return errors.New("reuseStore.PutPeopleToNotify requires SessionID")
	}

	names := map[string]string{}
	values := map[string]types.AttributeValue{}
	var statements []string
	for i, personToNotify := range peopleToNotify {
		index := strconv.Itoa(i)

		names["#ActorUID"+index] = personToNotify.UID.String()

		personToNotify.UID = actoruid.UID{}
		value, err := attributevalue.Marshal(personToNotify)
		if err != nil {
			return fmt.Errorf("marshal personToNotify: %w", err)
		}

		values[":Value"+index] = value

		statements = append(statements, "#ActorUID"+index+" = :Value"+index)
	}

	return s.dynamoClient.Update(ctx, dynamo.ReuseKey(data.SessionID, actor.TypePersonToNotify.String()), dynamo.MetadataKey(""),
		names,
		values,
		"SET "+strings.Join(statements, ", "),
	)
}

func (s *Store) DeletePersonToNotify(ctx context.Context, personToNotify donordata.PersonToNotify) error {
	return deleteReusable(ctx, s.dynamoClient, actor.TypePersonToNotify, personToNotify.UID)
}

func (s *Store) PeopleToNotify(ctx context.Context, provided *donordata.Provided) ([]donordata.PersonToNotify, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("reuseStore.PeopleToNotify requires SessionID")
	}

	var v map[string]orString[donordata.PersonToNotify]
	if err := s.dynamoClient.One(ctx, dynamo.ReuseKey(data.SessionID, actor.TypePersonToNotify.String()), dynamo.MetadataKey(""), &v); err != nil {
		return nil, err
	}

	delete(v, "PK")
	delete(v, "SK")

	seen := map[donordata.PersonToNotify]struct{}{}
	for _, personToNotify := range provided.PeopleToNotify {
		personToNotify.UID = actoruid.UID{}
		seen[personToNotify] = struct{}{}
	}

	var peopleToNotify []donordata.PersonToNotify
	for _, personToNotify := range v {
		if _, ok := seen[personToNotify.v]; !ok {
			peopleToNotify = append(peopleToNotify, personToNotify.v)
			seen[personToNotify.v] = struct{}{}
		}
	}

	slices.SortFunc(peopleToNotify, func(a, b donordata.PersonToNotify) int {
		return strings.Compare(a.FullName(), b.FullName())
	})

	return peopleToNotify, nil
}
