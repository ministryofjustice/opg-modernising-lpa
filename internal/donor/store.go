package donor

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dashboard/dashboarddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type DynamoClient interface {
	One(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error
	OneByPK(ctx context.Context, pk dynamo.PK, v interface{}) error
	OneByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	AllByPartialSK(ctx context.Context, pk dynamo.PK, partialSK dynamo.SK, v interface{}) error
	LatestForActor(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	AllByKeys(ctx context.Context, keys []dynamo.Keys) ([]map[string]dynamodbtypes.AttributeValue, error)
	AllKeysByPK(ctx context.Context, pk dynamo.PK) ([]dynamo.Keys, error)
	Put(ctx context.Context, v interface{}) error
	Create(ctx context.Context, v interface{}) error
	DeleteKeys(ctx context.Context, keys []dynamo.Keys) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	Update(ctx context.Context, pk dynamo.PK, sk dynamo.SK, values map[string]dynamodbtypes.AttributeValue, expression string) error
	BatchPut(ctx context.Context, items []interface{}) error
	OneBySK(ctx context.Context, sk dynamo.SK, v interface{}) error
	OneByUID(ctx context.Context, uid string, v interface{}) error
	WriteTransaction(ctx context.Context, transaction *dynamo.Transaction) error
}

type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (uid.CreateCaseResponse, error)
}

type EventClient interface {
	SendUidRequested(context.Context, event.UidRequested) error
	SendApplicationDeleted(context.Context, event.ApplicationDeleted) error
	SendApplicationUpdated(context.Context, event.ApplicationUpdated) error
	SendPreviousApplicationLinked(context.Context, event.PreviousApplicationLinked) error
	SendReducedFeeRequested(context.Context, event.ReducedFeeRequested) error
}

type SearchClient interface {
	Index(ctx context.Context, lpa search.Lpa) error
}

type Store struct {
	dynamoClient DynamoClient
	eventClient  EventClient
	logger       Logger
	uuidString   func() string
	newUID       func() actoruid.UID
	now          func() time.Time
	searchClient SearchClient
}

func NewStore(dynamoClient DynamoClient, eventClient EventClient, logger Logger, searchClient SearchClient) *Store {
	return &Store{
		dynamoClient: dynamoClient,
		eventClient:  eventClient,
		logger:       logger,
		uuidString:   uuid.NewString,
		newUID:       actoruid.New,
		now:          time.Now,
		searchClient: searchClient,
	}
}

func (s *Store) Create(ctx context.Context) (*donordata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Create requires SessionID")
	}

	lpaID := s.uuidString()
	donorUID := s.newUID()

	donor := &donordata.Provided{
		PK:        dynamo.LpaKey(lpaID),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(data.SessionID)),
		LpaID:     lpaID,
		CreatedAt: s.now(),
		Version:   1,
		Donor: donordata.Donor{
			UID:     donorUID,
			Channel: lpadata.ChannelOnline,
		},
		ProgressSteps: &task.Progress{},
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

	if err := donor.UpdateHash(); err != nil {
		return nil, err
	}

	if err := s.dynamoClient.Create(ctx, donor); err != nil {
		return nil, err
	}

	if err := s.dynamoClient.Create(ctx, dashboarddata.LpaLink{
		PK:        dynamo.LpaKey(lpaID),
		SK:        dynamo.SubKey(data.SessionID),
		DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey(data.SessionID)),
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
	PK           dynamo.LpaKeyType
	SK           dynamo.DonorKeyType
	ReferencedSK dynamo.OrganisationKeyType
}

// Link allows a donor to access an Lpa created by a supporter. It adds the donor's email to
// the share code and creates two records:
//
//  1. an lpaReference which allows the donor's session ID to be queried
//     for the organisation ID that holds the Lpa data;
//  2. an lpaLink which allows
//     the Lpa to be shown on the donor's dashboard.
func (s *Store) Link(ctx context.Context, shareCode sharecodedata.Link, donorEmail string) error {
	organisationKey, ok := shareCode.LpaOwnerKey.Organisation()
	if !ok {
		return errors.New("donorStore.Link can only be used with organisations")
	}

	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" {
		return errors.New("donorStore.Link requires SessionID")
	}

	var link dashboarddata.LpaLink
	if err := s.dynamoClient.OneByPartialSK(ctx, shareCode.LpaKey, dynamo.SubKey(""), &link); err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
		return err
	} else if link.ActorType == actor.TypeDonor {
		return errors.New("a donor link already exists for " + shareCode.LpaKey.ID())
	}

	shareCode.LpaLinkedTo = donorEmail
	shareCode.LpaLinkedAt = s.now()

	transaction := dynamo.NewTransaction().
		Create(lpaReference{
			PK:           shareCode.LpaKey,
			SK:           dynamo.DonorKey(data.SessionID),
			ReferencedSK: organisationKey,
		}).
		Create(dashboarddata.LpaLink{
			PK:        shareCode.LpaKey,
			SK:        dynamo.SubKey(data.SessionID),
			DonorKey:  shareCode.LpaOwnerKey,
			ActorType: actor.TypeDonor,
			UpdatedAt: s.now(),
		}).
		Put(shareCode)

	return s.dynamoClient.WriteTransaction(ctx, transaction)
}

func (s *Store) GetAny(ctx context.Context) (*donordata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("donorStore.GetAny requires LpaID")
	}

	var sk dynamo.SK = dynamo.DonorKey("")
	if data.OrganisationID != "" {
		sk = dynamo.OrganisationKey("")
	}

	var donor struct {
		donordata.Provided
		ReferencedSK dynamo.OrganisationKeyType
	}
	if err := s.dynamoClient.OneByPartialSK(ctx, dynamo.LpaKey(data.LpaID), sk, &donor); err != nil {
		return nil, err
	}

	if donor.ReferencedSK != "" {
		err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), donor.ReferencedSK, &donor)
	}

	return &donor.Provided, err
}

func (s *Store) Get(ctx context.Context) (*donordata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("donorStore.Get requires LpaID and SessionID")
	}

	var sk dynamo.SK = dynamo.DonorKey(data.SessionID)
	if data.OrganisationID != "" {
		sk = dynamo.OrganisationKey(data.OrganisationID)
	}

	var donor struct {
		donordata.Provided
		ReferencedSK dynamo.OrganisationKeyType
	}
	if err := s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), sk, &donor); err != nil {
		return nil, err
	}

	if donor.ReferencedSK != "" {
		err = s.dynamoClient.One(ctx, dynamo.LpaKey(data.LpaID), donor.ReferencedSK, &donor)
	}

	return &donor.Provided, err
}

func (s *Store) Latest(ctx context.Context) (*donordata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Latest requires SessionID")
	}

	var donor *donordata.Provided
	if err := s.dynamoClient.LatestForActor(ctx, dynamo.DonorKey(data.SessionID), &donor); err != nil {
		return nil, err
	}

	return donor, nil
}

func (s *Store) GetByKeys(ctx context.Context, keys []dynamo.Keys) ([]donordata.Provided, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	items, err := s.dynamoClient.AllByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}

	var donors []donordata.Provided
	err = attributevalue.UnmarshalListOfMaps(items, &donors)

	mappedDonors := map[string]donordata.Provided{}
	for _, donor := range donors {
		mappedDonors[donor.PK.PK()+"|"+donor.SK.SK()] = donor
	}

	clear(donors)
	for i, key := range keys {
		donors[i] = mappedDonors[key.PK.PK()+"|"+key.SK.SK()]
	}

	return donors, err
}

func (s *Store) Put(ctx context.Context, donor *donordata.Provided) error {
	if !donor.HashChanged() {
		return nil
	}

	// Enforces donor to send notifications to certificate provider when LPA data has changed
	if donor.CheckedHashChanged() && donor.Tasks.CheckYourLpa.IsCompleted() {
		donor.Tasks.CheckYourLpa = task.StateInProgress
	}

	if err := donor.UpdateHash(); err != nil {
		return err
	}

	// By not setting UpdatedAt until a UID exists, queries for SK=DONOR#xyz on
	// SKUpdatedAtIndex will not return UID-less LPAs.
	if donor.LpaUID != "" {
		donor.UpdatedAt = s.now()

		if err := s.searchClient.Index(ctx, search.Lpa{
			PK: donor.PK.PK(),
			SK: donor.SK.SK(),
			Donor: search.LpaDonor{
				FirstNames: donor.Donor.FirstNames,
				LastName:   donor.Donor.LastName,
			},
		}); err != nil {
			s.logger.WarnContext(ctx, "donorStore index failed", slog.Any("err", err))
		}
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

	return s.dynamoClient.Put(ctx, donor)
}

func (s *Store) Delete(ctx context.Context) error {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" || data.LpaID == "" {
		return errors.New("donorStore.Delete requires SessionID and LpaID")
	}

	keys, err := s.dynamoClient.AllKeysByPK(ctx, dynamo.LpaKey(data.LpaID))
	if err != nil {
		return err
	}

	canDelete := false
	for _, key := range keys {
		if key.PK == dynamo.LpaKey(data.LpaID) && key.SK == dynamo.DonorKey(data.SessionID) {
			canDelete = true
			break
		}
	}

	if !canDelete {
		return errors.New("cannot access data of another donor")
	}

	provided, err := s.Get(ctx)
	if err != nil {
		return err
	}

	if err = s.eventClient.SendApplicationDeleted(ctx, event.ApplicationDeleted{UID: provided.LpaUID}); err != nil {
		return err
	}

	return s.dynamoClient.DeleteKeys(ctx, keys)
}

func (s *Store) DeleteDonorAccess(ctx context.Context, shareCodeData sharecodedata.Link) error {
	organisationKey, ok := shareCodeData.LpaOwnerKey.Organisation()
	if !ok {
		return errors.New("donorStore.DeleteDonorAccess can only be used with organisations")
	}

	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return err
	}

	if data.OrganisationID == "" {
		return errors.New("donorStore.DeleteDonorAccess requires OrganisationID")
	}

	if data.OrganisationID != organisationKey.ID() {
		return errors.New("cannot remove access to another organisations LPA")
	}

	var link dashboarddata.LpaLink
	if err := s.dynamoClient.OneByPartialSK(ctx, shareCodeData.LpaKey, dynamo.SubKey(""), &link); err != nil {
		return err
	}

	transaction := dynamo.NewTransaction().
		Delete(dynamo.Keys{
			PK: link.PK,
			SK: link.SK,
		}).
		Delete(dynamo.Keys{
			PK: shareCodeData.LpaKey,
			SK: dynamo.DonorKey(link.UserSub()),
		}).
		Delete(dynamo.Keys{
			PK: shareCodeData.PK,
			SK: shareCodeData.SK,
		})

	return s.dynamoClient.WriteTransaction(ctx, transaction)
}
