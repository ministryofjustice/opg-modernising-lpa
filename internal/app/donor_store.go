package app

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

//go:generate mockery --testonly --inpackage --name UidClient --structname mockUidClient
type UidClient interface {
	CreateCase(context.Context, *uid.CreateCaseRequestBody) (uid.CreateCaseResponse, error)
}

//go:generate mockery --testonly --inpackage --name EventClient --structname mockEventClient
type EventClient interface {
	Send(context.Context, string, any) error
}

//go:generate mockery --testonly --inpackage --name DocumentStore --structname mockDocumentStore
type DocumentStore interface {
	GetAll(context.Context) (page.Documents, error)
	Put(context.Context, page.Document, []byte) error
	UpdateScanResults(context.Context, string, string, bool) error
}

type donorStore struct {
	dynamoClient  DynamoClient
	eventClient   EventClient
	uidClient     UidClient
	logger        Logger
	uuidString    func() string
	now           func() time.Time
	s3Client      *s3.Client
	documentStore DocumentStore
}

func (s *donorStore) Create(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Create requires SessionID")
	}

	lpaID := s.uuidString()

	lpa := &page.Lpa{
		PK:        lpaKey(lpaID),
		SK:        donorKey(data.SessionID),
		ID:        lpaID,
		CreatedAt: s.now(),
		Version:   1,
	}

	if err := s.dynamoClient.Create(ctx, lpa); err != nil {
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

	return lpa, err
}

func (s *donorStore) GetAny(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" {
		return nil, errors.New("donorStore.Get requires LpaID")
	}

	var lpa *page.Lpa
	if err := s.dynamoClient.OneByPartialSk(ctx, lpaKey(data.LpaID), "#DONOR#", &lpa); err != nil {
		return nil, err
	}

	return lpa, nil
}

func (s *donorStore) Get(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.LpaID == "" || data.SessionID == "" {
		return nil, errors.New("donorStore.Get requires LpaID and SessionID")
	}

	var lpa *page.Lpa
	err = s.dynamoClient.One(ctx, lpaKey(data.LpaID), donorKey(data.SessionID), &lpa)
	return lpa, err
}

func (s *donorStore) Latest(ctx context.Context) (*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.Get requires SessionID")
	}

	var lpa *page.Lpa
	if err := s.dynamoClient.LatestForActor(ctx, donorKey(data.SessionID), &lpa); err != nil {
		return nil, err
	}

	return lpa, nil
}

func (s *donorStore) Put(ctx context.Context, lpa *page.Lpa) error {
	if lpa.UID == "" && !lpa.Type.Empty() {
		resp, err := s.uidClient.CreateCase(ctx, &uid.CreateCaseRequestBody{
			Type: lpa.Type.String(),
			Donor: uid.DonorDetails{
				Name:     lpa.Donor.FullName(),
				Dob:      lpa.Donor.DateOfBirth,
				Postcode: lpa.Donor.Address.Postcode,
			},
		})
		if err != nil {
			s.logger.Print(err)
		} else {
			lpa.UID = resp.UID
		}
	}

	// By not setting UpdatedAt until a UID exists, queries for SK=#DONOR#xyz on
	// ActorUpdatedAtIndex will not return UID-less LPAs.
	if lpa.UID != "" {
		lpa.UpdatedAt = s.now()
	}

	if lpa.UID != "" && !lpa.HasSentApplicationUpdatedEvent {
		if err := s.eventClient.Send(ctx, "application-updated", applicationUpdatedEvent{
			UID:       lpa.UID,
			Type:      lpa.Type.String(),
			CreatedAt: lpa.CreatedAt,
			Donor: applicationUpdatedEventDonor{
				FirstNames:  lpa.Donor.FirstNames,
				LastName:    lpa.Donor.LastName,
				DateOfBirth: lpa.Donor.DateOfBirth,
				Postcode:    lpa.Donor.Address.Postcode,
			},
		}); err != nil {
			s.logger.Print(err)
		} else {
			lpa.HasSentApplicationUpdatedEvent = true
		}
	}

	if lpa.UID != "" && lpa.PreviousApplicationNumber != "" && !lpa.HasSentPreviousApplicationLinkedEvent {
		if err := s.eventClient.Send(ctx, "previous-application-linked", previousApplicationLinkedEvent{
			UID:                       lpa.UID,
			ApplicationReason:         lpa.ApplicationReason.String(),
			PreviousApplicationNumber: lpa.PreviousApplicationNumber,
		}); err != nil {
			s.logger.Print(err)
		} else {
			lpa.HasSentPreviousApplicationLinkedEvent = true
		}
	}

	if lpa.UID != "" && lpa.Tasks.PayForLpa.IsPending() {
		documents, err := s.documentStore.GetAll(ctx)
		if err != nil {
			return err
		}

		var unsentKeys []string

		for _, document := range documents {
			if document.Sent.IsZero() && !document.Scanned {
				unsentKeys = append(unsentKeys, document.Key)
			}
		}

		if len(unsentKeys) > 0 {
			if err := s.eventClient.Send(ctx, "reduced-fee-requested", reducedFeeRequestedEvent{
				UID:         lpa.UID,
				RequestType: lpa.FeeType.String(),
				Evidence:    unsentKeys,
			}); err != nil {
				s.logger.Print(err)
			} else {
				for _, document := range documents {
					if document.Sent.IsZero() && !document.Scanned {
						document.Sent = s.now()
						if err := s.documentStore.Put(ctx, document, nil); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return s.dynamoClient.Put(ctx, lpa)
}

func (s *donorStore) Delete(ctx context.Context) error {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return err
	}

	if data.SessionID == "" || data.LpaID == "" {
		return errors.New("donorStore.Create requires SessionID and LpaID")
	}

	keys, err := s.dynamoClient.AllKeysByPk(ctx, lpaKey(data.LpaID))
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

func lpaKey(s string) string {
	return "LPA#" + s
}

func donorKey(s string) string {
	return "#DONOR#" + s
}

func subKey(s string) string {
	return "#SUB#" + s
}

type applicationUpdatedEvent struct {
	UID       string                       `json:"uid"`
	Type      string                       `json:"type"`
	CreatedAt time.Time                    `json:"createdAt"`
	Donor     applicationUpdatedEventDonor `json:"donor"`
}

type applicationUpdatedEventDonor struct {
	FirstNames  string    `json:"firstNames"`
	LastName    string    `json:"lastName"`
	DateOfBirth date.Date `json:"dob"`
	Postcode    string    `json:"postcode"`
}

type previousApplicationLinkedEvent struct {
	UID                       string `json:"uid"`
	ApplicationReason         string `json:"applicationReason"`
	PreviousApplicationNumber string `json:"previousApplicationNumber"`
}

type reducedFeeRequestedEvent struct {
	UID         string   `json:"uid"`
	RequestType string   `json:"requestType"`
	Evidence    []string `json:"evidence"`
}
