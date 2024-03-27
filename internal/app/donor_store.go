package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (uid.CreateCaseResponse, error)
}

type EventClient interface {
	SendUidRequested(context.Context, event.UidRequested) error
	SendApplicationUpdated(context.Context, event.ApplicationUpdated) error
	SendPreviousApplicationLinked(context.Context, event.PreviousApplicationLinked) error
	SendReducedFeeRequested(context.Context, event.ReducedFeeRequested) error
}

type DocumentStore interface {
	GetAll(context.Context) (page.Documents, error)
	Put(context.Context, page.Document) error
	UpdateScanResults(context.Context, string, string, bool) error
}

type donorStore struct {
	dynamoClient  DynamoClient
	eventClient   EventClient
	logger        Logger
	uuidString    func() string
	newUID        func() actoruid.UID
	now           func() time.Time
	s3Client      *s3.Client
	documentStore DocumentStore
	searchClient  SearchClient
}

func (s *donorStore) Create(ctx context.Context) (*actor.DonorProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Create requires SessionID")
	}

	lpaID := s.uuidString()
	donorUID := s.newUID()

	donor := &actor.DonorProvidedDetails{
		PK:        lpaKey(lpaID),
		SK:        donorKey(data.SessionID),
		LpaID:     lpaID,
		CreatedAt: s.now(),
		Version:   1,
		Donor: actor.Donor{
			UID: donorUID,
		},
		Channel: actor.Online,
	}

	latest, err := s.Latest(ctx)
	if err != nil {
		return nil, err
	}

	if latest != nil {
		donor.Donor.FirstNames = latest.Donor.FirstNames
		donor.Donor.LastName = latest.Donor.LastName
		donor.Donor.OtherNames = latest.Donor.OtherNames
		donor.Donor.DateOfBirth = latest.Donor.DateOfBirth
		donor.Donor.Address = latest.Donor.Address
	}

	if donor.Hash, err = donor.GenerateHash(); err != nil {
		return nil, err
	}

	if err := s.dynamoClient.Create(ctx, donor); err != nil {
		return nil, err
	}

	if err := s.dynamoClient.Create(ctx, lpaLink{
		PK:        lpaKey(lpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(data.SessionID),
		ActorType: actor.TypeDonor,
		UpdatedAt: s.now(),
	}); err != nil {
		return nil, err
	}

	return donor, err
}

// An lpaReference creates a "pointer" record which can be queried as if the
// expected donor owned the LPA. This contains the actual SK containing the LPA
// data.
type lpaReference struct {
	PK, SK       string
	ReferencedSK string
}

// Link allows a donor to access an Lpa created by a supporter. It creates two
// records:
//
//  1. an lpaReference which allows the donor's session ID to be queried
//     for the organisation ID that holds the Lpa data;
//  2. an lpaLink which allows
//     the Lpa to be shown on the donor's dashboard.
func (s *donorStore) Link(ctx context.Context, shareCode actor.ShareCodeData) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("donorStore.Link requires SessionID")
	}

	var link lpaLink
	if err := s.dynamoClient.OneByPartialSK(ctx, lpaKey(shareCode.LpaID), subKey(""), &link); err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	} else if link.ActorType == actor.TypeDonor {
		return errors.New("a donor link already exists for " + shareCode.LpaID)
	}

	if err := s.dynamoClient.Create(ctx, lpaReference{
		PK:           lpaKey(shareCode.LpaID),
		SK:           donorKey(data.SessionID),
		ReferencedSK: organisationKey(shareCode.SessionID),
	}); err != nil {
		return err
	}

	return s.dynamoClient.Create(ctx, lpaLink{
		PK:        lpaKey(shareCode.LpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  organisationKey(shareCode.SessionID),
		ActorType: actor.TypeDonor,
		UpdatedAt: s.now(),
	})
}

func (s *donorStore) GetAny(ctx context.Context) (*actor.DonorProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("donorStore.Get requires LpaID")
	}

	var donor *actor.DonorProvidedDetails
	if err := s.dynamoClient.OneByPartialSK(ctx, lpaKey(data.LpaID), "#DONOR#", &donor); err != nil {
		return nil, err
	}

	return donor, nil
}

func (s *donorStore) Get(ctx context.Context) (*actor.DonorProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("donorStore.Get requires LpaID and SessionID")
	}

	sk := donorKey(data.SessionID)
	if data.OrganisationID != "" {
		sk = organisationKey(data.OrganisationID)
	}

	var donor struct {
		actor.DonorProvidedDetails
		ReferencedSK string
	}
	if err := s.dynamoClient.One(ctx, lpaKey(data.LpaID), sk, &donor); err != nil {
		return nil, err
	}

	if donor.ReferencedSK != "" {
		err = s.dynamoClient.One(ctx, lpaKey(data.LpaID), donor.ReferencedSK, &donor)
	}

	return &donor.DonorProvidedDetails, err
}

func (s *donorStore) Latest(ctx context.Context) (*actor.DonorProvidedDetails, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Get requires SessionID")
	}

	var donor *actor.DonorProvidedDetails
	if err := s.dynamoClient.LatestForActor(ctx, donorKey(data.SessionID), &donor); err != nil {
		return nil, err
	}

	return donor, nil
}

func (s *donorStore) GetByKeys(ctx context.Context, keys []dynamo.Key) ([]actor.DonorProvidedDetails, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	items, err := s.dynamoClient.AllByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}

	var donors []actor.DonorProvidedDetails
	err = attributevalue.UnmarshalListOfMaps(items, &donors)

	return donors, err
}

func (s *donorStore) Put(ctx context.Context, donor *actor.DonorProvidedDetails) error {
	newHash, err := donor.GenerateHash()
	if newHash == donor.Hash || err != nil {
		return err
	}

	donor.Hash = newHash

	// By not setting UpdatedAt until a UID exists, queries for SK=#DONOR#xyz on
	// SKUpdatedAtIndex will not return UID-less LPAs.
	if donor.LpaUID != "" {
		donor.UpdatedAt = s.now()

		if err := s.searchClient.Index(ctx, search.Lpa{
			PK:            donor.PK,
			SK:            donor.SK,
			DonorFullName: donor.Donor.FullName(),
		}); err != nil {
			return fmt.Errorf("donorStore index failed: %w", err)
		}
	}

	if donor.LpaUID == "" && !donor.Type.Empty() && !donor.HasSentUidRequestedEvent {
		data, err := page.SessionDataFromContext(ctx)
		if err != nil {
			return err
		}

		if err := s.eventClient.SendUidRequested(ctx, event.UidRequested{
			LpaID:          donor.LpaID,
			DonorSessionID: data.SessionID,
			OrganisationID: data.OrganisationID,
			Type:           donor.Type.String(),
			Donor: uid.DonorDetails{
				Name:     donor.Donor.FullName(),
				Dob:      donor.Donor.DateOfBirth,
				Postcode: donor.Donor.Address.Postcode,
			},
		}); err != nil {
			return err
		}

		donor.HasSentUidRequestedEvent = true
	}

	if donor.LpaUID != "" && !donor.HasSentApplicationUpdatedEvent {
		if err := s.eventClient.SendApplicationUpdated(ctx, event.ApplicationUpdated{
			UID:       donor.LpaUID,
			Type:      donor.Type.String(),
			CreatedAt: donor.CreatedAt,
			Donor: event.ApplicationUpdatedDonor{
				FirstNames:  donor.Donor.FirstNames,
				LastName:    donor.Donor.LastName,
				DateOfBirth: donor.Donor.DateOfBirth,
				Address:     donor.Donor.Address,
			},
		}); err != nil {
			return err
		}

		donor.HasSentApplicationUpdatedEvent = true
	}

	if donor.LpaUID != "" && donor.PreviousApplicationNumber != "" && !donor.HasSentPreviousApplicationLinkedEvent {
		if err := s.eventClient.SendPreviousApplicationLinked(ctx, event.PreviousApplicationLinked{
			UID:                       donor.LpaUID,
			PreviousApplicationNumber: donor.PreviousApplicationNumber,
		}); err != nil {
			return err
		}

		donor.HasSentPreviousApplicationLinkedEvent = true
	}

	return s.dynamoClient.Put(ctx, donor)
}

func (s *donorStore) Delete(ctx context.Context) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" || data.LpaID == "" {
		return errors.New("donorStore.Create requires SessionID and LpaID")
	}

	keys, err := s.dynamoClient.AllKeysByPK(ctx, lpaKey(data.LpaID))
	if err != nil {
		return err
	}

	canDelete := false
	for _, key := range keys {
		if key.PK == lpaKey(data.LpaID) && key.SK == donorKey(data.SessionID) {
			canDelete = true
			break
		}
	}

	if !canDelete {
		return errors.New("cannot access data of another donor")
	}

	return s.dynamoClient.DeleteKeys(ctx, keys)
}

func (s *donorStore) DeleteLink(ctx context.Context, shareCodeData actor.ShareCodeData) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID == "" {
		return errors.New("donorStore.DeleteLink requires OrganisationID")
	}

	if data.OrganisationID != shareCodeData.SessionID {
		return errors.New("cannot remove access to another organisations LPA")
	}

	var link lpaLink
	if err := s.dynamoClient.OneByPartialSK(ctx, lpaKey(shareCodeData.LpaID), subKey(""), &link); err != nil {
		return err
	}

	if err := s.dynamoClient.DeleteOne(ctx, link.PK, link.SK); err != nil {
		return err
	}

	return s.dynamoClient.DeleteOne(ctx, lpaKey(shareCodeData.LpaID), donorKey(link.UserSub()))
}

func lpaKey(s string) string {
	return "LPA#" + s
}

func donorKey(s string) string {
	return "#DONOR#" + s
}

func subKey(s string) string {
	return "#SUB#" + s
}
