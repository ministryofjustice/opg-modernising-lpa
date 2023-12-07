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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
)

const FormUrlEncoded = "application/x-www-form-urlencoded"

//go:generate mockery --testonly --inpackage --name Template --structname mockTemplate
type Template func(io.Writer, interface{}) error

//go:generate mockery --testonly --inpackage --name Logger --structname mockLogger
type Logger interface {
	Print(v ...interface{})
	Request(*http.Request, error)
}

//go:generate mockery --testonly --inpackage --name ShareCodeStore --structname mockShareCodeStore
type ShareCodeStore interface {
	Get(ctx context.Context, actorType actor.Type, shareCode string) (actor.ShareCodeData, error)
	Put(ctx context.Context, actorType actor.Type, shareCode string, data actor.ShareCodeData) error
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	SendEmail(context.Context, string, notify.Email) (string, error)
	SendSMS(context.Context, string, notify.SMS) (string, error)
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	EndSessionURL(idToken, postLogoutURL string) string
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

//go:generate mockery --testonly --inpackage --name DonorStore --structname mockDonorStore
type DonorStore interface {
	Create(context.Context) (*actor.DonorProvidedDetails, error)
	Put(context.Context, *actor.DonorProvidedDetails) error
}

//go:generate mockery --testonly --inpackage --name SessionStore --structname mockSessionStore
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	New(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

//go:generate mockery --testonly --inpackage --name Bundle --structname mockBundle
type Bundle interface {
	For(lang localize.Lang) *localize.Localizer
}

//go:generate mockery --testonly --inpackage --name Localizer --structname mockLocalizer
type Localizer interface {
	Format(string, map[string]any) string
	T(string) string
	Count(string, int) string
	FormatCount(string, int, map[string]interface{}) string
	ShowTranslationKeys() bool
	SetShowTranslationKeys(bool)
	Possessive(string) string
	Concat([]string, string) string
	FormatDate(date.TimeOrDate) string
	FormatDateTime(time.Time) string
}

//go:generate mockery --testonly --inpackage --name shareCodeSender --structname mockShareCodeSender
type shareCodeSender interface {
	SendCertificateProvider(ctx context.Context, appData AppData, identity bool, donor *actor.DonorProvidedDetails) error
	SendAttorneys(ctx context.Context, appData AppData, donor *actor.DonorProvidedDetails) error
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func PostFormReferenceNumber(r *http.Request, name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(name), " ", ""), "-", "")
}

//go:generate mockery --testonly --inpackage --name Handler --structname mockHandler
type Handler func(data AppData, w http.ResponseWriter, r *http.Request) error
