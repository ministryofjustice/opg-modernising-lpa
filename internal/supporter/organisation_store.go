package supporter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

type OrganisationStore struct {
	dynamoClient DynamoClient
	uuidString   func() string
	newUID       func() actoruid.UID
	randomString func(int) string
	now          func() time.Time
}

func NewOrganisationStore(dynamoClient DynamoClient) *OrganisationStore {
	return &OrganisationStore{
		dynamoClient: dynamoClient,
		uuidString:   random.UUID,
		newUID:       actoruid.New,
		randomString: random.AlphaNumeric,
		now:          time.Now,
	}
}

// An organisationLink is used to join a Member to an Organisation to be accessed by MemberID.
type organisationLink struct {
	// PK is the same as the PK for the Member
	PK dynamo.OrganisationKeyType
	// SK is the Member ID for the Member
	SK       dynamo.MemberIDKeyType
	MemberSK dynamo.MemberKeyType
}

func (s *OrganisationStore) Create(ctx context.Context, member *supporterdata.Member, name string) (*supporterdata.Organisation, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Create requires SessionID")
	}

	organisation := &supporterdata.Organisation{
		PK:        dynamo.OrganisationKey(member.OrganisationID),
		SK:        dynamo.OrganisationKey(member.OrganisationID),
		ID:        member.OrganisationID,
		Name:      name,
		CreatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, organisation); err != nil {
		return nil, fmt.Errorf("error creating organisation: %w", err)
	}

	return organisation, nil
}

func (s *OrganisationStore) Get(ctx context.Context) (*supporterdata.Organisation, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("organisationStore.Get requires SessionID")
	}

	var member supporterdata.Member
	if err := s.dynamoClient.OneBySK(ctx, dynamo.MemberKey(data.SessionID), &member); err != nil {
		return nil, err
	}

	var organisation supporterdata.Organisation
	if err := s.dynamoClient.One(ctx, member.PK, member.PK, &organisation); err != nil {
		return nil, err
	}

	if !organisation.DeletedAt.IsZero() {
		return nil, dynamo.NotFoundError{}
	}

	return &organisation, err
}

func (s *OrganisationStore) Put(ctx context.Context, organisation *supporterdata.Organisation) error {
	organisation.UpdatedAt = s.now()
	return s.dynamoClient.Put(ctx, organisation)
}

func (s *OrganisationStore) CreateLPA(ctx context.Context) (*donordata.Provided, error) {
	data, err := appcontext.SessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.OrganisationID == "" {
		return nil, errors.New("organisationStore.CreateLPA requires OrganisationID")
	}

	lpaID := s.uuidString()
	donorUID := s.newUID()

	donor := &donordata.Provided{
		PK:        dynamo.LpaKey(lpaID),
		SK:        dynamo.LpaOwnerKey(dynamo.OrganisationKey(data.OrganisationID)),
		LpaID:     lpaID,
		CreatedAt: s.now(),
		Version:   1,
		Donor: donordata.Donor{
			UID: donorUID,
		},
	}

	if err := donor.UpdateHash(); err != nil {
		return nil, err
	}

	transaction := dynamo.NewTransaction().
		Create(dynamo.Keys{PK: donor.PK, SK: dynamo.ReservedKey(dynamo.DonorKey)}).
		Create(donor)

	if err := s.dynamoClient.WriteTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	return donor, err
}

func (s *OrganisationStore) SoftDelete(ctx context.Context, organisation *supporterdata.Organisation) error {
	organisation.DeletedAt = s.now()

	return s.dynamoClient.Put(ctx, organisation)
}
