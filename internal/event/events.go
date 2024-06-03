package event

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type UidRequested struct {
	LpaID          string
	DonorSessionID string
	OrganisationID string
	Type           string
	Donor          uid.DonorDetails
}

type ApplicationUpdated struct {
	UID       string                  `json:"uid"`
	Type      string                  `json:"type"`
	CreatedAt time.Time               `json:"createdAt"`
	Donor     ApplicationUpdatedDonor `json:"donor"`
}

type ApplicationUpdatedDonor struct {
	FirstNames  string        `json:"firstNames"`
	LastName    string        `json:"lastName"`
	DateOfBirth date.Date     `json:"dob"`
	Address     place.Address `json:"address"`
}

type PreviousApplicationLinked struct {
	UID                       string `json:"uid"`
	PreviousApplicationNumber string `json:"previousApplicationNumber"`
}

type ReducedFeeRequested struct {
	UID              string     `json:"uid"`
	RequestType      string     `json:"requestType"`
	Evidence         []Evidence `json:"evidence,omitempty"`
	EvidenceDelivery string     `json:"evidenceDelivery"`
}

type Evidence struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
}

type NotificationSent struct {
	UID            string `json:"uid"`
	NotificationID string `json:"notificationId"`
}

type PaperFormRequested struct {
	UID        string       `json:"uid"`
	ActorType  string       `json:"actorType"`
	ActorUID   actoruid.UID `json:"actorUID"`
	AccessCode string       `json:"accessCode"`
}

type PaymentReceived struct {
	UID       string `json:"uid"`
	PaymentID string `json:"paymentId"`
	Amount    int    `json:"amount"`
}
