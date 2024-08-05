package supporterdata

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMemberStoreCreateMemberInvite(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.MemberInvite{
			PK:               dynamo.OrganisationKey("an-id"),
			SK:               dynamo.MemberInviteKey("email@example.com"),
			CreatedAt:        testNow,
			OrganisationID:   "a-uuid",
			OrganisationName: "org name",
			Email:            "email@example.com",
			FirstNames:       "a",
			LastName:         "b",
			Permission:       actor.PermissionNone,
			ReferenceNumber:  "abcde",
		}).
		Return(nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn}

	err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{ID: "a-uuid", Name: "org name"}, "a", "b", "email@example.com", "abcde", actor.PermissionNone)
	assert.Nil(t, err)
}

func TestMemberStoreCreateMemberInviteWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"missing session":        context.Background(),
		"missing OrganisationID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
	}

	memberStore := &MemberStore{now: testNowFn}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {

			err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.PermissionNone)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreInvitedMember(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneBySK(ctx, dynamo.MemberInviteKey("a@example.org"), &actor.MemberInvite{OrganisationID: "an-id"}, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMember, err := memberStore.InvitedMember(ctx)

	assert.Nil(t, err)
	assert.Equal(t, &actor.MemberInvite{OrganisationID: "an-id"}, invitedMember)
}

func TestMemberStoreInvitedMemberWhenDynamoError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneBySK(ctx, mock.Anything, mock.Anything, expectedError)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.InvitedMember(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreInvitedMemberWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no email":        appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.InvitedMember(ctx)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreCreateMemberInviteWhenErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn}

	err := memberStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.PermissionNone)
	assert.ErrorIs(t, err, expectedError)
}

func TestMemberStoreInvitedMembers(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, dynamo.OrganisationKey("an-id"),
		dynamo.MemberInviteKey(""), []*actor.MemberInvite{{OrganisationID: "an-id"}, {OrganisationID: "an-id"}}, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMembers, err := memberStore.InvitedMembers(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.MemberInvite{{OrganisationID: "an-id"}, {OrganisationID: "an-id"}}, invitedMembers)
}

func TestMemberStoreInvitedMembersWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no organisation id": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.InvitedMembers(ctx)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreInvitedMembersWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, dynamo.OrganisationKey("an-id"),
		dynamo.MemberInviteKey(""), []*actor.MemberInvite{}, expectedError)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.InvitedMembers(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreInvitedMembersByEmail(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.MemberInviteKey("a@example.org"), []*actor.MemberInvite{
		{OrganisationID: "an-id", Email: "a@example.org"},
		{OrganisationID: "another-id", Email: "a@example.org"},
	}, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMembers, err := memberStore.InvitedMembersByEmail(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.MemberInvite{
		{OrganisationID: "an-id", Email: "a@example.org"},
		{OrganisationID: "another-id", Email: "a@example.org"},
	}, invitedMembers)
}

func TestMemberStoreInvitedMembersByEmailWhenDynamoError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, mock.Anything, mock.Anything, expectedError)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.InvitedMembersByEmail(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreInvitedMembersByEmailWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no email":        appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.InvitedMembersByEmail(ctx)

			assert.Error(t, err)
		})
	}
}

func TestPut(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.Member{PK: dynamo.OrganisationKey("123"), SK: dynamo.MemberKey("456"), UpdatedAt: testNow}).
		Return(nil)

	store := &MemberStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Member{PK: dynamo.OrganisationKey("123"), SK: dynamo.MemberKey("456")})
	assert.Nil(t, err)
}

func TestPutWhenDynamoError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &MemberStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Member{PK: dynamo.OrganisationKey("123"), SK: dynamo.MemberKey("456")})
	assert.Equal(t, expectedError, err)
}

func TestMemberStoreGetAll(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, dynamo.OrganisationKey("an-id"),
		dynamo.MemberKey(""), []*actor.Member{{FirstNames: "a"}, {FirstNames: "c"}, {FirstNames: "b"}}, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	members, err := memberStore.GetAll(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.Member{{FirstNames: "a"}, {FirstNames: "b"}, {FirstNames: "c"}}, members)
}

func TestMemberStoreGetAllWhenSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no organisation ID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.GetAll(ctx)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreGetAllWhenDynamoClientError(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSK(ctx, dynamo.OrganisationKey("an-id"),
		dynamo.MemberKey(""), []*actor.MemberInvite{}, expectedError)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.GetAll(ctx)

	assert.Equal(t, expectedError, err)
}

func TestMemberStoreGet(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{
		OrganisationID: "a-uuid",
		SessionID:      "session-id",
	})
	member := &actor.Member{FirstNames: "a"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.OrganisationKey("a-uuid"), dynamo.MemberKey("session-id"),
			member, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := memberStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, member, result)
}

func TestMemberStoreGetWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":      appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "id"}),
		"no organisation id": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{}

			_, err := memberStore.Get(ctx)
			assert.Error(t, err)
		})
	}
}

func TestMemberStoreGetWhenErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "a-uuid", SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.OrganisationKey("a-uuid"), dynamo.MemberKey("session-id"),
			nil, expectedError)
	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.Get(ctx)
	assert.Equal(t, expectedError, err)
}

func TestMemberStoreCreate(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id", Email: "email@example.com"})
	expectedMember := &actor.Member{
		PK:             dynamo.OrganisationKey("a-uuid"),
		SK:             dynamo.MemberKey("session-id"),
		CreatedAt:      testNow,
		UpdatedAt:      testNow,
		ID:             "a-uuid",
		OrganisationID: "a-uuid",
		Email:          "email@example.com",
		FirstNames:     "a",
		LastName:       "b",
		Permission:     actor.PermissionAdmin,
		Status:         actor.StatusActive,
		LastLoggedInAt: testNow,
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, expectedMember).
		Return(nil).
		Once()
	dynamoClient.EXPECT().
		Create(ctx, &organisationLink{
			PK:       dynamo.OrganisationKey("a-uuid"),
			SK:       dynamo.MemberIDKey("a-uuid"),
			MemberSK: dynamo.MemberKey("session-id"),
		}).
		Return(nil).
		Once()

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	member, err := memberStore.Create(ctx, "a", "b")
	assert.Nil(t, err)
	assert.Equal(t, expectedMember, member)
}

func TestMemberStoreCreateWhenSessionMissing(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":    context.Background(),
		"missing email":      appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "a"}),
		"missing session ID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{Email: "a"}),
	}

	for name, ctx := range testCases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.Create(ctx, "a", "b")
			assert.Error(t, err)
		})
	}
}

func TestMemberStoreCreateWhenDynamoErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id", Email: "a"})

	testcases := map[string]struct {
		dynamoClientSetup func(*mockDynamoClient)
	}{
		"member": {
			dynamoClientSetup: func(dynamoClient *mockDynamoClient) {
				dynamoClient.EXPECT().
					Create(ctx, mock.Anything).
					Return(expectedError).
					Once()
			},
		},
		"link": {
			dynamoClientSetup: func(dynamoClient *mockDynamoClient) {
				dynamoClient.EXPECT().
					Create(ctx, mock.Anything).
					Return(nil).
					Once()
				dynamoClient.EXPECT().
					Create(ctx, mock.Anything).
					Return(expectedError).
					Once()
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			tc.dynamoClientSetup(dynamoClient)

			memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.Create(ctx, "a", "b")
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestMemberStoreCreateFromInvite(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	invite := &actor.MemberInvite{
		PK:             "pk",
		SK:             "sk",
		Email:          "ab@example.org",
		FirstNames:     "a",
		LastName:       "b",
		Permission:     actor.PermissionAdmin,
		OrganisationID: "org-id",
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Member{
			PK:             dynamo.OrganisationKey("org-id"),
			SK:             dynamo.MemberKey("session-id"),
			CreatedAt:      testNow,
			UpdatedAt:      testNow,
			ID:             "a-uuid",
			OrganisationID: "org-id",
			Email:          invite.Email,
			FirstNames:     invite.FirstNames,
			LastName:       invite.LastName,
			Permission:     invite.Permission,
			Status:         actor.StatusActive,
			LastLoggedInAt: testNow,
		}).
		Return(nil)

	dynamoClient.EXPECT().
		DeleteOne(ctx, dynamo.OrganisationKey("org-id"), dynamo.MemberInviteKey(invite.Email)).
		Return(nil)

	dynamoClient.EXPECT().
		Create(ctx, &organisationLink{
			PK:       dynamo.OrganisationKey("org-id"),
			SK:       dynamo.MemberIDKey("a-uuid"),
			MemberSK: dynamo.MemberKey("session-id"),
		}).
		Return(nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	err := memberStore.CreateFromInvite(ctx, invite)
	assert.Nil(t, err)
}

func TestMemberStoreCreateFromInviteWhenSessionMissing(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":    context.Background(),
		"missing session ID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
	}

	for name, ctx := range testCases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			err := memberStore.CreateFromInvite(ctx, &actor.MemberInvite{})
			assert.Error(t, err)
		})
	}
}

func TestMemberStoreCreateFromInviteWhenDynamoErrors(t *testing.T) {
	testcases := map[string]struct {
		createMemberError error
		deleteOneError    error
		createLinkError   error
	}{
		"Create member error": {
			createMemberError: expectedError,
		},
		"DeleteOne error": {
			deleteOneError: expectedError,
		},
		"Create link error": {
			createLinkError: expectedError,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(tc.createMemberError).
				Once()

			if tc.deleteOneError != nil || tc.createLinkError != nil {
				dynamoClient.EXPECT().
					DeleteOne(ctx, mock.Anything, mock.Anything).
					Return(tc.deleteOneError)
			}

			if tc.createLinkError != nil {
				dynamoClient.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(tc.createLinkError).
					Once()
			}
			memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			err := memberStore.CreateFromInvite(ctx, &actor.MemberInvite{})
			assert.Error(t, err)
		})
	}
}

func TestMemberStoreGetByID(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "org-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, dynamo.OrganisationKey("org-id"), dynamo.MemberIDKey("1"),
			&organisationLink{
				PK:       dynamo.OrganisationKey("org-id"),
				SK:       dynamo.MemberIDKey("1"),
				MemberSK: dynamo.MemberKey("a-uuid"),
			}, nil)

	expectedMember := &actor.Member{
		PK: dynamo.OrganisationKey("org-id"),
		SK: dynamo.MemberKey("a-uuid"),
		ID: "1",
	}

	dynamoClient.
		ExpectOne(ctx, dynamo.OrganisationKey("org-id"), dynamo.MemberKey("a-uuid"),
			expectedMember, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	member, err := memberStore.GetByID(ctx, "1")

	assert.Nil(t, err)
	assert.Equal(t, expectedMember, member)
}

func TestMemberStoreGetByIDWhenMissingSession(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":         context.Background(),
		"missing organisation ID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
	}

	for name, ctx := range testCases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.GetByID(ctx, "1")

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreGetByIDWhenDynamoClientErrors(t *testing.T) {
	testCases := map[string]struct {
		oneLinkError   error
		oneMemberError error
	}{
		"one link error": {
			oneLinkError: expectedError,
		},
		"one member error": {
			oneMemberError: expectedError,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{OrganisationID: "org-id"})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOne(mock.Anything, mock.Anything, mock.Anything,
					&organisationLink{}, tc.oneLinkError)

			if tc.oneMemberError != nil {
				dynamoClient.
					ExpectOne(mock.Anything, mock.Anything, mock.Anything,
						&actor.Member{}, tc.oneMemberError)
			}

			memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.GetByID(ctx, "1")

			assert.Equal(t, expectedError, err)
		})
	}
}

func TestMemberStoreGetAny(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	expectedMember := &actor.Member{ID: "a"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.MemberKey("session-id"),
			expectedMember, nil)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	member, err := memberStore.GetAny(ctx)
	assert.Nil(t, err)
	assert.Equal(t, expectedMember, member)
}

func TestMemberStoreGetAnyWhenMissingSession(t *testing.T) {
	testCases := map[string]context.Context{
		"missing session":    context.Background(),
		"missing session ID": appcontext.ContextWithSession(context.Background(), &appcontext.Session{}),
	}

	for name, ctx := range testCases {
		t.Run(name, func(t *testing.T) {
			memberStore := &MemberStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := memberStore.GetAny(ctx)

			assert.Error(t, err)
		})
	}
}

func TestMemberStoreGetAnyWhenDynamoClientErrors(t *testing.T) {
	ctx := appcontext.ContextWithSession(context.Background(), &appcontext.Session{SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, dynamo.MemberKey("session-id"),
			nil, expectedError)

	memberStore := &MemberStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := memberStore.GetAny(ctx)

	assert.Equal(t, expectedError, err)
}
