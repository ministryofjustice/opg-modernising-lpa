package supporterdata

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

const memberInviteExpireAfter = time.Hour * 48

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
