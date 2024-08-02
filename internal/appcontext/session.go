package appcontext

import "context"

type Session struct {
	SessionID string
	LpaID     string

	// if a supporter
	Email          string
	OrganisationID string
}

type SessionMissingError struct{}

func (s SessionMissingError) Error() string {
	return "session not set in context"
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	data, ok := ctx.Value((*Session)(nil)).(*Session)

	if !ok {
		return nil, SessionMissingError{}
	}

	return data, nil
}

func ContextWithSession(ctx context.Context, data *Session) context.Context {
	return context.WithValue(ctx, (*Session)(nil), data)
}
