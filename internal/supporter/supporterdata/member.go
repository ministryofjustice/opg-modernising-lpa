package supporterdata

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

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
