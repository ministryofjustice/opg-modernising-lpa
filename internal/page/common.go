package page

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
)

const FormUrlEncoded = "application/x-www-form-urlencoded"

type Template func(io.Writer, interface{}) error

type Logger interface {
	Print(v ...interface{})
	Request(*http.Request, error)
}

type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (actor.ShareCodeData, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error
}

type NotifyClient interface {
	SendActorEmail(context context.Context, to, lpaUID string, email notify.Email) error
	SendActorSMS(context context.Context, to, lpaUID string, sms notify.SMS) error
}

type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) (string, error)
	EndSessionURL(idToken, postLogoutURL string) (string, error)
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

type DonorStore interface {
	Create(context.Context) (*actor.DonorProvidedDetails, error)
	Put(context.Context, *actor.DonorProvidedDetails) error
}

type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

type Bundle interface {
	For(lang localize.Lang) *localize.Localizer
}

type Localizer interface {
	Concat([]string, string) string
	Count(string, int) string
	Format(string, map[string]any) string
	FormatCount(string, int, map[string]interface{}) string
	FormatDate(date.TimeOrDate) string
	FormatDateTime(time.Time) string
	Possessive(string) string
	SetShowTranslationKeys(bool)
	ShowTranslationKeys() bool
	T(string) string
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func PostFormReferenceNumber(r *http.Request, name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(name), " ", ""), "-", "")
}

type Handler func(data AppData, w http.ResponseWriter, r *http.Request) error

type EventClient interface {
	SendNotificationSent(context.Context, event.NotificationSent) error
}
