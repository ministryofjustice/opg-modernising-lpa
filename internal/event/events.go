package event

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type UidRequested struct {
	LpaID          string           `json:"lpaID"`
	DonorSessionID string           `json:"donorSessionID"`
	OrganisationID string           `json:"organisationID"`
	Type           string           `json:"type"`
	Donor          uid.DonorDetails `json:"donor"`
}

type ApplicationDeleted struct {
	UID string `json:"uid"`
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

type CertificateProviderStarted struct {
	UID string `json:"uid"`
}

type AttorneyStarted struct {
	LpaUID   string       `json:"uid"`
	ActorUID actoruid.UID `json:"actorUID"`
}

type IdentityCheckMismatched struct {
	LpaUID   string                         `json:"uid"`
	ActorUID actoruid.UID                   `json:"actorUID"`
	Provided IdentityCheckMismatchedDetails `json:"provided"`
	Verified IdentityCheckMismatchedDetails `json:"verified"`
}

type IdentityCheckMismatchedDetails struct {
	FirstNames  string    `json:"firstNames"`
	LastName    string    `json:"lastName"`
	DateOfBirth date.Date `json:"dateOfBirth"`
}

type CorrespondentUpdated struct {
	UID        string         `json:"uid"`
	FirstNames string         `json:"firstNames,omitempty"`
	LastName   string         `json:"lastName,omitempty"`
	Email      string         `json:"email,omitempty"`
	Phone      string         `json:"phone,omitempty"`
	Address    *place.Address `json:"address,omitempty"`
}

type LpaAccessGranted struct {
	UID     string                  `json:"uid"`
	LpaType string                  `json:"lpaType"`
	Actors  []LpaAccessGrantedActor `json:"actors"`
}

type LpaAccessGrantedActor struct {
	ActorUID  string `json:"actorUid"`
	SubjectID string `json:"subjectId"`
}
