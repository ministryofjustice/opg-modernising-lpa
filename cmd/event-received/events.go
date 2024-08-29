package main

import "time"

type uidEvent struct {
	UID string `json:"uid"`
}

type furtherInfoRequestedEvent struct {
	UID        string    `json:"uid"`
	PostedDate time.Time `json:"postedDate"`
}

type lpaUpdatedEvent struct {
	UID        string `json:"uid"`
	ChangeType string `json:"changeType"`
}
