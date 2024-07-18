package page

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

const FormUrlEncoded = "application/x-www-form-urlencoded"

type Template func(io.Writer, interface{}) error

type Logger interface {
	InfoContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
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

type Bundle interface {
	For(lang localize.Lang) *localize.Localizer
}

type Localizer interface {
	Concat(list []string, joiner string) string
	Count(messageID string, count int) string
	Format(messageID string, data map[string]interface{}) string
	FormatCount(messageID string, count int, data map[string]any) string
	FormatDate(t date.TimeOrDate) string
	FormatTime(t time.Time) string
	FormatDateTime(t time.Time) string
	Possessive(s string) string
	SetShowTranslationKeys(s bool)
	ShowTranslationKeys() bool
	T(messageID string) string
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func PostFormReferenceNumber(r *http.Request, name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(name), " ", ""), "-", "")
}

type Handler func(data AppData, w http.ResponseWriter, r *http.Request) error

type EventClient interface {
	SendNotificationSent(ctx context.Context, notificationSentEvent event.NotificationSent) error
	SendPaperFormRequested(ctx context.Context, paperFormRequestedEvent event.PaperFormRequested) error
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
