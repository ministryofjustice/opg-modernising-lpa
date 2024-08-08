package lpadata

import "time"

type IdentityCheck struct {
	CheckedAt time.Time `json:"checkedAt"`
	Type      string    `json:"type"`
}
