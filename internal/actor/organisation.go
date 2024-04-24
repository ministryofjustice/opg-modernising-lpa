package actor

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

const memberInviteExpireAfter = time.Hour * 48

// An Organisation contains users associated with a set of permissions that work on the
// same set of LPAs.
type Organisation struct {
	PK, SK dynamo.OrganisationKeyType
	// CreatedAt is when the Organisation was created
	CreatedAt time.Time
	// UpdatedAt is when the Organisation was last updated
	UpdatedAt time.Time
	// DeletedAt is when the Organisation was (soft) deleted
	DeletedAt time.Time
	// ID is a unique identifier for the Organisation
	ID string
	// Name of the Organisation, this is unique across all Organisations
	Name string
}

// A Member is the association of a OneLogin user with an Organisation.
type Member struct {
	PK dynamo.OrganisationKeyType
	SK dynamo.MemberKeyType
	// CreatedAt is when the Member was created
	CreatedAt time.Time
	// UpdatedAt is when the Member was last updated
	UpdatedAt time.Time
	// ID is a unique identifier for the Member
	ID string
	// OrganisationID identifies the organisation the member belongs to
	OrganisationID string
	Email          string
	FirstNames     string
	LastName       string
	// Permission is the type of permissions assigned to the member to set available actions in an Organisation
	Permission Permission
	// Status controls access to the Organisation
	Status Status
	// LastLoggedInAt is when the Member last logged in to the service
	LastLoggedInAt time.Time
}

func (i Member) FullName() string {
	return fmt.Sprintf("%s %s", i.FirstNames, i.LastName)
}

// A MemberInvite is created to allow a new Member to join an Organisation
type MemberInvite struct {
	PK dynamo.OrganisationKeyType
	SK dynamo.MemberInviteKeyType
	// CreatedAt is when the MemberInvite was created
	CreatedAt time.Time
	// UpdatedAt is when the MemberInvite was last updated
	UpdatedAt time.Time
	// OrganisationID identifies the organisation the invite is for
	OrganisationID string
	// OrganisationName is the name of the organisation the invite is for
	OrganisationName string
	// Email is the address the new Member must sign in as for the invite
	Email      string
	FirstNames string
	LastName   string
	// Permission is the type of permissions assigned to the member to set available actions in an Organisation
	Permission Permission
	// ReferenceNumber is a unique code used to invite a Member to and Organisation
	ReferenceNumber string
}

func (i MemberInvite) HasExpired() bool {
	return i.CreatedAt.Add(memberInviteExpireAfter).Before(time.Now())
}

func (i MemberInvite) FullName() string {
	return fmt.Sprintf("%s %s", i.FirstNames, i.LastName)
}
