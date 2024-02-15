package app

import (
	"context"
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestOrganisationStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Organisation{
			PK:        "ORGANISATION#a-uuid",
			SK:        "ORGANISATION#a-uuid",
			ID:        "a-uuid",
			CreatedAt: testNow,
			Name:      "A name",
		}).
		Return(nil)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Member{
			PK:        "ORGANISATION#a-uuid",
			SK:        "MEMBER#an-id",
			CreatedAt: testNow,
			Email:     "a@example.org",
		}).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	organisation, err := organisationStore.Create(ctx, "A name")
	assert.Nil(t, err)
	assert.Equal(t, &actor.Organisation{
		PK:        "ORGANISATION#a-uuid",
		SK:        "ORGANISATION#a-uuid",
		ID:        "a-uuid",
		CreatedAt: testNow,
		Name:      "A name",
	}, organisation)
}

func TestOrganisationStoreCreateWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":   page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "a@example.org"}),
		"no email":        page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			organisation, err := organisationStore.Create(ctx, "A name")
			assert.Error(t, err)
			assert.Nil(t, organisation)
		})
	}
}

func TestOrganisationStoreCreateWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id", Email: "a@example.org"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"organisation": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"member": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDynamoClient(t)
			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			organisation, err := organisationStore.Create(ctx, "A name")
			assert.ErrorIs(t, err, expectedError)
			assert.Nil(t, organisation)
		})
	}
}

func TestOrganisationStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	organisation := &actor.Organisation{Name: "A name"}

	member := actor.Member{PK: "ORGANISATION#a-uuid"}
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneBySK(ctx, "MEMBER#session-id", member, nil)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "ORGANISATION#a-uuid", organisation, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := organisationStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, organisation, result)
}

func TestOrganisationStoreGetWithSessionMissing(t *testing.T) {
	organisationStore := &organisationStore{}

	_, err := organisationStore.Get(context.Background())
	assert.Error(t, err)
}

func TestOrganisationStoreGetWhenErrors(t *testing.T) {
	testcases := map[string]struct {
		oneBySKError error
		oneError     error
	}{
		"OneBySK error": {
			oneBySKError: expectedError,
		},
		"One error": {
			oneError: expectedError,
		},
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "session-id"})
	member := actor.Member{PK: "ORGANISATION#a-uuid"}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneBySK(ctx, "MEMBER#session-id", member, tc.oneBySKError)

			if tc.oneError != nil {
				dynamoClient.
					ExpectOne(ctx, "ORGANISATION#a-uuid", "ORGANISATION#a-uuid", nil, tc.oneError)
			}

			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := organisationStore.Get(ctx)
			assert.Equal(t, expectedError, err)
		})
	}
}

func TestOrganisationStorePut(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.Organisation{PK: "ORGANISATION#123", SK: "ORGANISATION#456", Name: "Hey", UpdatedAt: testNow}).
		Return(expectedError)

	store := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.Put(ctx, &actor.Organisation{PK: "ORGANISATION#123", SK: "ORGANISATION#456", Name: "Hey"})
	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreCreateMemberInvite(t *testing.T) {
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

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn}

	err := organisationStore.CreateMemberInvite(ctx, &actor.Organisation{ID: "a-uuid", Name: "org name"}, "a", "b", "email@example.com", "abcde", actor.None)
	assert.Nil(t, err)
}

func TestOrganisationStoreCreateMemberInviteWhenMissingOrganisationID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	organisationStore := &organisationStore{now: testNowFn}

	err := organisationStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.None)
	assert.Equal(t, errors.New("organisationStore.Get requires OrganisationID"), err)
}

func TestOrganisationStoreCreateMemberInviteWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn}

	err := organisationStore.CreateMemberInvite(ctx, &actor.Organisation{}, "a", "b", "email@example.com", "abcde", actor.None)
	assert.ErrorIs(t, err, expectedError)
}

func TestOrganisationStoreCreateLPA(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})
	expectedDonor := &actor.DonorProvidedDetails{
		PK:        "LPA#a-uuid",
		SK:        "ORGANISATION#an-id",
		LpaID:     "a-uuid",
		CreatedAt: testNow,
		Version:   1,
	}
	expectedDonor.Hash, _ = expectedDonor.GenerateHash()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, expectedDonor).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	donor, err := organisationStore.CreateLPA(ctx)

	assert.Nil(t, err)
	assert.Equal(t, expectedDonor, donor)
}

func TestOrganisationStoreCreateLPAWithSessionMissing(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: ""})

	organisationStore := &organisationStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.CreateLPA(ctx)

	assert.NotNil(t, err)
}

func TestOrganisationStoreCreateLPAMissingOrganisationID(t *testing.T) {
	ctx := context.Background()

	organisationStore := &organisationStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.CreateLPA(ctx)

	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestOrganisationStoreCreateLPAWhenDynamoError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, mock.Anything).
		Return(expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.CreateLPA(ctx)

	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreAllLPAs(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})
	expectedDonorA := actor.DonorProvidedDetails{
		PK:     "LPA#a-uuid",
		SK:     "ORGANISATION#an-id",
		LpaUID: "a-uid",
		Donor: actor.Donor{
			FirstNames: "a",
			LastName:   "a",
		},
	}
	expectedDonorB := actor.DonorProvidedDetails{
		PK:     "LPA#b-uuid",
		SK:     "ORGANISATION#an-id",
		LpaUID: "b-uid",
		Donor: actor.Donor{
			FirstNames: "a",
			LastName:   "b",
		},
	}
	expectedDonorC := actor.DonorProvidedDetails{
		PK:     "LPA#c-uuid",
		SK:     "ORGANISATION#an-id",
		LpaUID: "c-uid",
		Donor: actor.Donor{
			FirstNames: "c",
			LastName:   "a",
		},
	}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, "ORGANISATION#an-id",
		[]actor.DonorProvidedDetails{
			expectedDonorB,
			expectedDonorC,
			expectedDonorA,
			{PK: "ORGANISATION#an-id", SK: "ORGANISATION#an-id"},
			{
				PK:    "LPA#d-uuid",
				SK:    "ORGANISATION#an-id",
				LpaID: "d-uuid",
				Donor: actor.Donor{
					FirstNames: "d",
					LastName:   "d",
				},
			},
		}, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	donors, err := organisationStore.AllLPAs(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []actor.DonorProvidedDetails{expectedDonorA, expectedDonorB, expectedDonorC}, donors)
}

func TestOrganisationStoreAllLPAsWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":   page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			donors, err := organisationStore.AllLPAs(ctx)
			assert.Error(t, err)
			assert.Nil(t, donors)
		})
	}
}

func TestOrganisationStoreAllLPAsWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, "ORGANISATION#an-id",
		nil, expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.AllLPAs(ctx)
	assert.ErrorIs(t, err, expectedError)
}

func TestOrganisationStoreInvitedMembers(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSk(ctx, "ORGANISATION#an-id",
		"MEMBERINVITE#", []*actor.MemberInvite{{OrganisationID: "an-id"}, {OrganisationID: "an-id"}}, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMembers, err := organisationStore.InvitedMembers(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.MemberInvite{{OrganisationID: "an-id"}, {OrganisationID: "an-id"}}, invitedMembers)
}

func TestOrganisationStoreInvitedMembersWhenSessionMissingOrgID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	organisationStore := &organisationStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.InvitedMembers(ctx)

	assert.Equal(t, errors.New("organisationStore.InvitedMembers requires OrganisationID"), err)
}

func TestOrganisationStoreInvitedMembersWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSk(ctx, "ORGANISATION#an-id",
		"MEMBERINVITE#", []*actor.MemberInvite{}, expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.InvitedMembers(ctx)

	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreInvitedMember(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneBySK(ctx, "MEMBERINVITE#YUBleGFtcGxlLm9yZw==", &actor.MemberInvite{OrganisationID: "an-id"}, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	invitedMember, err := organisationStore.InvitedMember(ctx)

	assert.Nil(t, err)
	assert.Equal(t, &actor.MemberInvite{OrganisationID: "an-id"}, invitedMember)
}

func TestOrganisationStoreInvitedMemberWhenDynamoError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{Email: "a@example.org"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectOneBySK(ctx, mock.Anything, mock.Anything, expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.InvitedMember(ctx)

	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreInvitedMemberWhenMissingEmail(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	organisationStore := &organisationStore{dynamoClient: nil, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.InvitedMember(ctx)

	assert.Equal(t, errors.New("organisationStore.InvitedMember requires Email"), err)
}

func TestPutMember(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(ctx, &actor.Member{PK: "ORGANISATION#123", SK: "ORGANISATION#456", UpdatedAt: testNow}).
		Return(nil)

	store := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.PutMember(ctx, &actor.Member{PK: "ORGANISATION#123", SK: "ORGANISATION#456"})
	assert.Nil(t, err)
}

func TestPutMemberWhenDynamoError(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &organisationStore{
		dynamoClient: dynamoClient,
		now:          testNowFn,
	}

	err := store.PutMember(ctx, &actor.Member{PK: "ORGANISATION#123", SK: "ORGANISATION#456"})
	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreMembers(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSk(ctx, "ORGANISATION#an-id",
		"MEMBER#", []*actor.Member{{FirstNames: "a"}, {FirstNames: "b"}}, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	members, err := organisationStore.Members(ctx)

	assert.Nil(t, err)
	assert.Equal(t, []*actor.Member{{FirstNames: "a"}, {FirstNames: "b"}}, members)
}

func TestOrganisationStoreMembersWhenSessionMissingOrgID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})

	organisationStore := &organisationStore{now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.Members(ctx)

	assert.Equal(t, errors.New("organisationStore.Members requires OrganisationID"), err)
}

func TestOrganisationStoreMembersWhenDynamoClientError(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllByPartialSk(ctx, "ORGANISATION#an-id",
		"MEMBER#", []*actor.MemberInvite{}, expectedError)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.Members(ctx)

	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreMember(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{
		OrganisationID: "a-uuid",
		SessionID:      "session-id",
	})
	member := &actor.Member{FirstNames: "a"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "MEMBER#session-id",
			member, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := organisationStore.Member(ctx)
	assert.Nil(t, err)
	assert.Equal(t, member, result)
}

func TestOrganisationStoreMemberWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":      page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "id"}),
		"no organisation id": page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data":    context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			_, err := organisationStore.Member(ctx)
			assert.Error(t, err)
		})
	}
}

func TestOrganisationStoreMemberWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{OrganisationID: "a-uuid", SessionID: "session-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "MEMBER#session-id",
			nil, expectedError)
	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	_, err := organisationStore.Member(ctx)
	assert.Equal(t, expectedError, err)
}

func TestOrganisationStoreCreateMember(t *testing.T) {
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
			Email:      invite.Email,
			FirstNames: invite.FirstNames,
			LastName:   invite.LastName,
			Permission: invite.Permission,
		}).
		Return(nil)

	dynamoClient.EXPECT().
		DeleteOne(ctx, "pk", "sk").
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	err := organisationStore.CreateMember(ctx, invite)
	assert.Nil(t, err)
}

func TestOrganisationStoreCreateMemberWhenDynamoErrors(t *testing.T) {
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

			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			err := organisationStore.CreateMember(ctx, &actor.MemberInvite{})
			assert.Error(t, err)
		})
	}

}
