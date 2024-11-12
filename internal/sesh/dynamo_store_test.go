package sesh

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func (c *mockDynamoClient_One_Call) SetData(data any) {
	c.Run(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func TestNewDynamoStore(t *testing.T) {
	client := newMockDynamoClient(t)
	store := NewDynamoStore(client)

	assert.Equal(t, &DynamoStore{
		Codecs: securecookie.CodecsFromPairs(),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		dynamoClient: client,
	}, store)
}

func TestDynamoStoreGet(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	dynamoClient := newMockDynamoClient(t)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Options:      &sessions.Options{Path: "/"},
	}

	session, err := store.Get(r, "name")
	assert.Nil(t, err)
	assert.True(t, session.IsNew)
	assert.Equal(t, &sessions.Options{Path: "/"}, session.Options)
}

func TestDynamoStoreNew(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/path?a=b", nil)

	dynamoClient := newMockDynamoClient(t)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Options:      &sessions.Options{Path: "/"},
	}

	session, err := store.New(r, "name")
	assert.Nil(t, err)
	assert.True(t, session.IsNew)
	assert.Equal(t, &sessions.Options{Path: "/"}, session.Options)
}

func TestDynamoStoreNewWhenExisting(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)

		sessionID      = "a-random-session-id"
		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/"}
		data           = map[any]any{"session": &LoginSession{Sub: "x"}}
	)

	codecs := securecookie.CodecsFromPairs(securecookie.GenerateRandomKey(32))
	encodedCookie, _ := codecs[0].Encode(sessionName, sessionID)
	encodedData, _ := codecs[0].Encode(sessionName, data)

	r.AddCookie(&http.Cookie{
		Name:  sessionName,
		Value: encodedCookie,
	})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(r.Context(), dynamo.SessionKey(sessionID), dynamo.MetadataKey(sessionID), mock.Anything).
		Return(nil).
		SetData(sessionData{Encoded: encodedData})

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Codecs:       codecs,
		Options:      sessionOptions,
	}

	session, err := store.New(r, sessionName)
	assert.Nil(t, err)
	assert.False(t, session.IsNew)
	assert.Equal(t, sessionOptions, session.Options)
	assert.Equal(t, data, session.Values)
}

func TestDynamoStoreNewWhenDynamoErrors(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)

		sessionID      = "a-random-session-id"
		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/"}
	)

	codecs := securecookie.CodecsFromPairs(securecookie.GenerateRandomKey(32))
	encodedCookie, _ := codecs[0].Encode(sessionName, sessionID)

	r.AddCookie(&http.Cookie{
		Name:  sessionName,
		Value: encodedCookie,
	})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Codecs:       codecs,
		Options:      sessionOptions,
	}

	session, err := store.New(r, sessionName)
	assert.Equal(t, expectedError, err)
	assert.True(t, session.IsNew)
	assert.Equal(t, sessionOptions, session.Options)
}

func TestDynamoStoreSave(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()

		sessionID      = "a-random-session-id"
		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/", MaxAge: 30}
		data           = map[any]any{"session": &LoginSession{Sub: "x"}}
	)

	codecs := securecookie.CodecsFromPairs(securecookie.GenerateRandomKey(32))
	encodedCookie, _ := codecs[0].Encode(sessionName, sessionID)
	encodedData, _ := codecs[0].Encode(sessionName, data)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(r.Context(), sessionData{
			PK:      dynamo.SessionKey(sessionID),
			SK:      dynamo.MetadataKey(sessionID),
			Encoded: encodedData,
		}).
		Return(nil)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Options:      sessionOptions,
		Codecs:       codecs,
	}

	session := sessions.NewSession(store, sessionName)
	session.ID = sessionID
	session.Values = data
	session.Options = sessionOptions

	err := store.Save(r, w, session)
	assert.Nil(t, err)

	cookie := w.Result().Cookies()[0]
	assert.Equal(t, sessionName, cookie.Name)
	assert.Equal(t, encodedCookie, cookie.Value)
	assert.WithinDuration(t, time.Now().Add(time.Duration(sessionOptions.MaxAge)*time.Second), cookie.Expires, time.Second)
}

func TestDynamoStoreSaveWhenDynamoErrors(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()

		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/", MaxAge: 30}
		codecs         = securecookie.CodecsFromPairs(securecookie.GenerateRandomKey(32))
	)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(expectedError)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Options:      sessionOptions,
		Codecs:       codecs,
	}

	session := sessions.NewSession(store, sessionName)
	session.Options = sessionOptions

	err := store.Save(r, w, session)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, w.Result().Cookies())
}

func TestDynamoStoreSaveWhenCodecErrors(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()

		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/", MaxAge: 30}
		data           = map[any]any{"session": &LoginSession{Sub: "x"}}
	)

	store := &DynamoStore{
		Options: sessionOptions,
		Codecs:  securecookie.CodecsFromPairs(),
	}

	session := sessions.NewSession(store, sessionName)
	session.Values = data
	session.Options = sessionOptions

	err := store.Save(r, w, session)
	assert.NotNil(t, err)
	assert.Empty(t, w.Result().Cookies())
}

func TestDynamoStoreSaveWhenExpiring(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()

		sessionID      = "a-random-session-id"
		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/", MaxAge: -1}
		data           = map[any]any{"session": &LoginSession{Sub: "x"}}
	)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(r.Context(), dynamo.SessionKey(sessionID), dynamo.MetadataKey(sessionID)).
		Return(nil)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
		Options:      sessionOptions,
	}

	session := sessions.NewSession(store, sessionName)
	session.ID = sessionID
	session.Values = data
	session.Options = sessionOptions

	err := store.Save(r, w, session)
	assert.Nil(t, err)

	cookie := w.Result().Cookies()[0]
	assert.Equal(t, sessionName, cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, time.Unix(1, 0).UTC(), cookie.Expires)
}

func TestDynamoStoreSaveWhenExpiringErrors(t *testing.T) {
	var (
		r, _ = http.NewRequest(http.MethodGet, "/path?a=b", nil)
		w    = httptest.NewRecorder()

		sessionName    = "a-session-name"
		sessionOptions = &sessions.Options{Path: "/", MaxAge: -1}
	)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		DeleteOne(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	store := &DynamoStore{
		dynamoClient: dynamoClient,
	}

	session := sessions.NewSession(store, sessionName)
	session.Options = sessionOptions

	err := store.Save(r, w, session)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, w.Result().Cookies())
}
