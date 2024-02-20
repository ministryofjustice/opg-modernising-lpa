package app

import (
	"context"
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMemberStoreCreateMemberInvite(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.MemberInvite{
			PK:               "ORGANISATION#an-id",
			SK:               "MEMBERINVITE#ZW1haWxAZXhhbXBsZS5jb20=",
			CreatedAt:        testNow,
			OrganisationID:   "a-uuid",
			OrganisationName: "org name",
			Email:            "email@example.com",
			FirstNames:       "a",
			LastName:         "b",
			Permission:       actor.None,
			ReferenceNumber:  "abcde",
		}).
		Return(nil)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn}

	err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{ID: "a-uuid", Name: "org name"}, "a", "b", "email@example.com", "abcde", actor.None)
	assert.Nil(t, err)
}

func TestMemberStoreCreateMemberInviteWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"missing session":        context.Background(),
		"missing OrganisationID": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
	}

	memberStore := &memberStore{now: testNowFn}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {

			err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.None)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreCreateMemberInviteWhenMissingOrganisationID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	memberStore := &memberStore{now: testNowFn}

	err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.None)
	assert.Equal(t, errors.New("memberStore.Get requires OrganisationID"), err)
}

func TestMemberStoreCreateMemberInviteWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn}

	err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.None)
	assert.ErrorIs(t, err, expectedError)
}

func TestMemberStoreInvitedMembers(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, "ORGANISATION#an-id",
		"MEMBERINVITE#", []*actor.MemberInvite{{OrganisationID: "an-id"}, {OrganisationID: "an-id"}}, nil)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMembers, err := memberStore.InvitedMembers(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.MemberInvite{{OrganisationID: "an-id"}, {OrganisationID: "an-id"}}, invitedMembers)
}

func TestMemberStoreInvitedMembersWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no organisation id": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &memberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.InvitedMembers(ctx)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreInvitedMembersWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, "ORGANISATION#an-id",
		"MEMBERINVITE#", []*actor.MemberInvite{}, expectedError)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.InvitedMembers(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreInvitedMember(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneBySK(ctx, "MEMBERINVITE#YUBleGFtcGxlLm9yZw==", &actor.MemberInvite{OrganisationID: "an-id"}, nil)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMember, err := memberStore.InvitedMember(ctx)

	assert.Nil(t, err)
	assert.Equal(t, &actor.MemberInvite{OrganisationID: "an-id"}, invitedMember)
}

func TestMemberStoreInvitedMemberWhenDynamoError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneBySK(ctx, mock.Anything, mock.Anything, expectedError)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.InvitedMember(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreInvitedMemberWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no email":        page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &memberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.InvitedMember(ctx)

			assert.Error(t, err)
		})
	}
}

func TestPut(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.Member{PK: "ORGANISATION#123", SK: "ORGANISATION#456", UpdatedAt: testNow}).
		Return(nil)

	store := &memberStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Member{PK: "ORGANISATION#123", SK: "ORGANISATION#456"})
	assert.Nil(t, err)
}

func TestPutWhenDynamoError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &memberStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Member{PK: "ORGANISATION#123", SK: "ORGANISATION#456"})
	assert.Equal(t, expectedError, err)
}

func TestMemberStoreMembers(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, "ORGANISATION#an-id",
		"MEMBER#", []*actor.Member{{FirstNames: "a"}, {FirstNames: "b"}}, nil)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	members, err := memberStore.Members(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.Member{{FirstNames: "a"}, {FirstNames: "b"}}, members)
}

func TestMemberStoreMembersWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no organisation ID": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &memberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.Members(ctx)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreMembersWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, "ORGANISATION#an-id",
		"MEMBER#", []*actor.MemberInvite{}, expectedError)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.Members(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreMember(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{
		OrganisationID: "a-uuid",
		SessionID:      "session-id",
	})
	member := &actor.Member{FirstNames: "a"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "MEMBER#session-id",
			member, nil)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := memberStore.Self(ctx)
	assert.Nil(t, err)
	assert.Equal(t, member, result)
}

func TestMemberStoreMemberWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":      page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "id"}),
		"no organisation id": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &memberStore{}

			_, err := memberStore.Self(ctx)
			assert.Error(t, err)
		})
	}
}

func TestMemberStoreMemberWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "a-uuid", SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "MEMBER#session-id",
			nil, expectedError)
	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.Self(ctx)
	assert.Equal(t, expectedError, err)
}

func TestMemberStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})

	invite := &actor.MemberInvite{
		PK:             "pk",
		SK:             "sk",
		Email:          "ab@example.ord",
		FirstNames:     "a",
		LastName:       "b",
		Permission:     actor.Admin,
		OrganisationID: "org-id",
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Member{
			PK:         "ORGANISATION#org-id",
			SK:         "MEMBER#session-id",
			CreatedAt:  testNow,
			UpdatedAt:  testNow,
			ID:         "a-uuid",
			Email:      invite.Email,
			FirstNames: invite.FirstNames,
			LastName:   invite.LastName,
			Permission: invite.Permission,
		}).
		Return(nil)

	dynamoClient.EXPECT().
		DeleteOne(ctx, "pk", "sk").
		Return(nil)

	dynamoClient.EXPECT().
		Create(ctx, &organisationLink{
			PK:       "ORGANISATION#org-id",
			SK:       "MEMBERID#a-uuid",
			MemberSK: "MEMBER#session-id",
		}).
		Return(nil)

	memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	err := memberStore.Create(ctx, invite)
	assert.Nil(t, err)
}

func TestMemberStoreCreateWhenSessionMissing(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":    context.Background(),
		"missing session ID": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
	}

	for name, ctx := range testCases {
		t.Run(name, func(t *testing.T) {
			memberStore := &memberStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			err := memberStore.Create(ctx, &actor.MemberInvite{})
			assert.Error(t, err)
		})
	}
}

func TestMemberStoreCreateWhenDynamoErrors(t *testing.T) {
	testcases := map[string]struct {
		createError    error
		deleteOneError error
	}{
		"Create error":    {createError: expectedError},
		"DeleteOne error": {deleteOneError: expectedError},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(tc.createError)

			if tc.deleteOneError != nil {
				dynamoClient.EXPECT().
					DeleteOne(ctx, mock.Anything, mock.Anything).
					Return(tc.deleteOneError)
			}

			memberStore := &memberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			err := memberStore.Create(ctx, &actor.MemberInvite{})
			assert.Error(t, err)
		})
	}

}
