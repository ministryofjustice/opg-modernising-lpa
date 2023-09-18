package app

import (
	"context"
	"errors"
	"slices"
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

type donorStore struct {
	dynamoClient DynamoClient
	eventClient  EventClient
	uidClient    UidClient
	logger       Logger
	uuidString   func() string
	now          func() time.Time
	s3Client     *s3.Client
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
		UpdatedAt: s.now(),
	}

	if err := s.dynamoClient.Create(ctx, lpa); err != nil {
		return nil, err
	}
	if err := s.dynamoClient.Create(ctx, lpaLink{
		PK:        lpaKey(lpaID),
		SK:        subKey(data.SessionID),
		DonorKey:  donorKey(data.SessionID),
		ActorType: actor.TypeDonor,
	}); err != nil {
		return nil, err
	}

	return lpa, err
}

func (s *donorStore) GetAll(ctx context.Context) ([]*page.Lpa, error) {
	data, err := page.SessionDataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if data.SessionID == "" {
		return nil, errors.New("donorStore.GetAll requires SessionID")
	}

	var items []*page.Lpa
	err = s.dynamoClient.GetAllByGsi(ctx, "ActorIndex", donorKey(data.SessionID), &items)

	slices.SortFunc(items, func(a, b *page.Lpa) int {
		if a.UpdatedAt.After(b.UpdatedAt) {
			return -1
		}
		return 1
	})

	return items, err
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
	if err := s.dynamoClient.GetOneByPartialSk(ctx, lpaKey(data.LpaID), "#DONOR#", &lpa); err != nil {
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
	err = s.dynamoClient.Get(ctx, lpaKey(data.LpaID), donorKey(data.SessionID), &lpa)
	return lpa, err
}

func (s *donorStore) Put(ctx context.Context, lpa *page.Lpa) error {
	lpa.UpdatedAt = s.now()

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

	if lpa.UID != "" && lpa.EvidenceFormAddress.Line1 != "" && !lpa.HasSentEvidenceFormRequiredEvent {
		if err := s.eventClient.Send(ctx, "evidence-form-required", evidenceFormRequiredEvent{
			UID:        lpa.UID,
			FirstNames: lpa.Donor.FirstNames,
			LastName:   lpa.Donor.LastName,
			Address: address{
				Line1:      lpa.EvidenceFormAddress.Line1,
				Line2:      lpa.EvidenceFormAddress.Line2,
				Line3:      lpa.EvidenceFormAddress.Line3,
				TownOrCity: lpa.EvidenceFormAddress.TownOrCity,
				Postcode:   lpa.EvidenceFormAddress.Postcode,
			},
		}); err != nil {
			s.logger.Print(err)
		} else {
			lpa.HasSentEvidenceFormRequiredEvent = true
		}
	}

	if lpa.UID != "" && lpa.Tasks.PayForLpa.IsPending() && lpa.HasUnsentReducedFeesEvidence() {
		var unsentKeys []string

		for _, evidence := range lpa.EvidenceKeys {
			if evidence.Sent.IsZero() {
				unsentKeys = append(unsentKeys, evidence.Key)
			}
		}

		if err := s.eventClient.Send(ctx, "reduced-fee-requested", reducedFeeRequestedEvent{
			UID:         lpa.UID,
			RequestType: lpa.FeeType.String(),
			Evidence:    unsentKeys,
		}); err != nil {
			s.logger.Print(err)
		} else {
			for i, evidence := range lpa.EvidenceKeys {
				if evidence.Sent.IsZero() {
					lpa.EvidenceKeys[i].Sent = s.now()
				}
			}
		}
	}

	return s.dynamoClient.Put(ctx, lpa)
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

type evidenceFormRequiredEvent struct {
	UID        string  `json:"uid"`
	FirstNames string  `json:"firstNames"`
	LastName   string  `json:"lastName"`
	Address    address `json:"address"`
}

type reducedFeeRequestedEvent struct {
	UID         string   `json:"uid"`
	RequestType string   `json:"requestType"`
	Evidence    []string `json:"evidence"`
}

type address struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	Line3      string `json:"line3,omitempty"`
	TownOrCity string `json:"townOrCity,omitempty"`
	Postcode   string `json:"postcode,omitempty"`
}
