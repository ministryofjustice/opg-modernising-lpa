package sesh

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type DynamoClient interface {
	Create(ctx context.Context, v any) error
	DeleteOne(ctx context.Context, pk dynamo.PK, sk dynamo.SK) error
	OneActive(ctx context.Context, pk dynamo.PK, sk dynamo.SK, now time.Time, v interface{}) error
}

func NewDynamoStore(dynamoClient DynamoClient, now func() time.Time, keyPairs ...[]byte) *DynamoStore {
	return &DynamoStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		dynamoClient: dynamoClient,
		now:          now,
	}
}

// DynamoStore stores sessions in DynamoDB.
type DynamoStore struct {
	Codecs       []securecookie.Codec
	Options      *sessions.Options
	dynamoClient DynamoClient
	now          func() time.Time
}

func (s *DynamoStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *DynamoStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(r.Context(), session)
			if err == nil {
				session.IsNew = false
			}
		}
	}

	return session, err
}

func (s *DynamoStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		if err := s.erase(r.Context(), session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = base64.RawStdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	}
	if err := s.save(r.Context(), session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

type sessionData struct {
	PK        dynamo.SessionKeyType
	SK        dynamo.MetadataKeyType
	Encoded   string
	ExpiresAt int64
}

func (s *DynamoStore) save(ctx context.Context, session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}

	expiresAt := s.now().UTC().Add(time.Duration(session.Options.MaxAge) * time.Second)

	return s.dynamoClient.Create(ctx, sessionData{
		PK:        dynamo.SessionKey(session.ID),
		SK:        dynamo.MetadataKey(session.ID),
		Encoded:   encoded,
		ExpiresAt: expiresAt.Unix(),
	})
}

func (s *DynamoStore) load(ctx context.Context, session *sessions.Session) error {
	var v sessionData
	if err := s.dynamoClient.OneActive(ctx, dynamo.SessionKey(session.ID), dynamo.MetadataKey(session.ID), s.now(), &v); err != nil {
		return err
	}

	return securecookie.DecodeMulti(session.Name(), v.Encoded, &session.Values, s.Codecs...)
}

func (s *DynamoStore) erase(ctx context.Context, session *sessions.Session) error {
	return s.dynamoClient.DeleteOne(ctx,
		dynamo.SessionKey(session.ID),
		dynamo.MetadataKey(session.ID),
	)
}
