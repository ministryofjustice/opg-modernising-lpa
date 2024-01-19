package actor

import "time"

// A Group contains users associated with a set of permissions that work on the
// same set of LPAs.
type Group struct {
	PK, SK string
	// CreatedAt is when the Group was created
	CreatedAt time.Time
	// UpdatedAt is when the Group was last updated
	UpdatedAt time.Time
	// ID is a unique identifier for the group
	ID string
	// Name of the group, this is unique across all groups
	Name string
}

// A GroupMember is the association of a OneLogin user with a Group.
type GroupMember struct {
	PK, SK string
	// CreatedAt is when the GroupMember was created
	CreatedAt time.Time
	// UpdatedAt is when the GroupMember was last updated
	UpdatedAt time.Time
}
