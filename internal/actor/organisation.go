package actor

import (
	"strings"
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
}

func (m *Member) OrganisationID() string {
	return strings.Split(m.SK, "ORGANISATION#")[1]
}

// A MemberInvite is created to allow a new Member to join an Organisation
type MemberInvite struct {
	PK, SK string
	// CreatedAt is when the MemberInvite was created
	CreatedAt time.Time
	// OrganisationID identifies the organisation the invite is for
	OrganisationID string
	// Email is the address the new Member must signin as for the invite
	Email string
}

func (i MemberInvite) HasExpired() bool {
	return i.CreatedAt.Add(memberInviteExpireAfter).Before(time.Now())
}
