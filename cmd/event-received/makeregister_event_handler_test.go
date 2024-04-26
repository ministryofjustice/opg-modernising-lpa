package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMakeRegisterHandlerHandleUnknownEvent(t *testing.T) {
	handler := &makeregisterEventHandler{}

	err := handler.Handle(ctx, nil, events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown cloudwatch event"), err)
}

func TestHandleUidRequested(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","organisationID":"org-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, &uid.CreateCaseRequestBody{
			Type: "personal-welfare",
			Donor: uid.DonorDetails{
				Name:     "a donor",
				Dob:      date.New("2000", "01", "02"),
				Postcode: "F1 1FF",
			},
		}).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, "an-id", "donor-id", "org-id", "M-1111-2222-3333").
		Return(nil)

	err := handleUidRequested(ctx, uidStore, uidClient, event)
	assert.Nil(t, err)
}

func TestHandleUidRequestedWhenUidClientErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("", expectedError)

	err := handleUidRequested(ctx, nil, uidClient, event)
	assert.Equal(t, fmt.Errorf("failed to create case: %w", expectedError), err)
}

func TestHandleUidRequestedWhenUidStoreErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, "an-id", "donor-id", "", "M-1111-2222-3333").
		Return(expectedError)

	err := handleUidRequested(ctx, uidStore, uidClient, event)
	assert.Equal(t, fmt.Errorf("failed to set uid: %w", expectedError), err)
}
