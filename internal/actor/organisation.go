package actor

import (
	"fmt"
	"time"
)

const memberInviteExpireAfter = time.Hour * 48

// An Organisation contains users associated with a set of permissions that work on the
// same set of LPAs.
type Organisation struct {
	PK, SK string
	// CreatedAt is when the Organisation was created
	CreatedAt time.Time
	// UpdatedAt is when the Organisation was last updated
	UpdatedAt time.Time
	// ID is a unique identifier for the Organisation
	ID string
	// Name of the Organisation, this is unique across all Organisations
	Name string
}

// A Member is the association of a OneLogin user with an Organisation.
type Member struct {
	PK, SK string
	// CreatedAt is when the Member was created
	CreatedAt time.Time
	// UpdatedAt is when the Member was last updated
	UpdatedAt time.Time
	// LastLoggedInAt is when the Member last logged in to the service
	LastLoggedInAt time.Time
	Email          string
	FirstNames     string
	LastName       string
	// Permission is the type of permissions assigned to the member to set available actions in an Organisation
	Permission Permission
}

func (i Member) FullName() string {
	return fmt.Sprintf("%s %s", i.FirstNames, i.LastName)
}

// A MemberInvite is created to allow a new Member to join an Organisation
type MemberInvite struct {
	PK, SK string
	// CreatedAt is when the MemberInvite was created
	CreatedAt time.Time
	// UpdatedAt is when the MemberInvite was last updated
	UpdatedAt time.Time
	// OrganisationID identifies the organisation the invite is for
	OrganisationID string
	// Email is the address the new Member must signin as for the invite
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
