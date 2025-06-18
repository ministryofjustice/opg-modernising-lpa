package page

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

const FormUrlEncoded = "application/x-www-form-urlencoded"

type Template func(io.Writer, interface{}) error

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, confidenceLevel onelogin.ConfidenceLevel) (string, error)
	EndSessionURL(idToken, postLogoutURL string) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type DonorStore interface {
	Create(context.Context) (*donordata.Provided, error)
	Put(context.Context, *donordata.Provided) error
}

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

type Localizer interface {
	localize.Localizer
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func PostFormReferenceNumber(r *http.Request, name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(name), " ", ""), "-", "")
}

type Handler func(data appcontext.Data, w http.ResponseWriter, r *http.Request) error

type EventClient interface {
	SendNotificationSent(ctx context.Context, notificationSentEvent event.NotificationSent) error
	SendPaperFormRequested(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested) error
	SendMetric(ctx context.Context, category event.Category, measure event.Measure) error
}

type SessionStore interface {
	Csrf(r *http.Request) (*sesh.CsrfSession, error)
	SetCsrf(r *http.Request, w http.ResponseWriter, session *sesh.CsrfSession) error
	Login(r *http.Request) (*sesh.LoginSession, error)
	SetLogin(r *http.Request, w http.ResponseWriter, session *sesh.LoginSession) error
	ClearLogin(r *http.Request, w http.ResponseWriter) error
	OneLogin(r *http.Request) (*sesh.OneLoginSession, error)
	SetOneLogin(r *http.Request, w http.ResponseWriter, session *sesh.OneLoginSession) error
}

type Notification struct {
	Heading  string
	BodyHTML string
}
