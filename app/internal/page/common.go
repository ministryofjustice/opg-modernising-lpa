package page

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
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

//go:generate mockery --testonly --inpackage --name DataStore --structname mockDataStore
type DataStore interface {
	Get(ctx context.Context, pk, sk string, v interface{}) error
	Put(context.Context, string, string, interface{}) error
	GetOneByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error
	GetAllByGsi(ctx context.Context, gsi, sk string, v interface{}) error
}

//go:generate mockery --testonly --inpackage --name NotifyClient --structname mockNotifyClient
type NotifyClient interface {
	Email(ctx context.Context, email notify.Email) (string, error)
	Sms(ctx context.Context, sms notify.Sms) (string, error)
	TemplateID(id notify.TemplateId) string
}

//go:generate mockery --testonly --inpackage --name OneLoginClient --structname mockOneLoginClient
type OneLoginClient interface {
	AuthCodeURL(state, nonce, locale string, identity bool) string
	Exchange(ctx context.Context, code, nonce string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
	ParseIdentityClaim(ctx context.Context, userInfo onelogin.UserInfo) (identity.UserData, error)
}

//go:generate mockery --testonly --inpackage --name LpaStore --structname mockLpaStore
type LpaStore interface {
	Create(context.Context) (*Lpa, error)
	GetAll(context.Context) ([]*Lpa, error)
	Get(context.Context) (*Lpa, error)
	Put(context.Context, *Lpa) error
}

//go:generate mockery --testonly --inpackage --name CertificateProviderStore --structname mockCertificateProviderStore
type CertificateProviderStore interface {
	Create(ctx context.Context) (*actor.CertificateProvider, error)
	Get(ctx context.Context) (*actor.CertificateProvider, error)
	Put(ctx context.Context, certificateProvider *actor.CertificateProvider) error
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
	Count(messageID string, count int) string
	FormatCount(messageID string, count int, data map[string]interface{}) string
	ShowTranslationKeys() bool
	SetShowTranslationKeys(s bool)
	Possessive(s string) string
}

//go:generate mockery --testonly --inpackage --name shareCodeSender --structname mockShareCodeSender
type shareCodeSender interface {
	SendCertificateProvider(ctx context.Context, template notify.TemplateId, appData AppData, identity bool, lpa *Lpa) error
	SendAttorneys(ctx context.Context, template notify.TemplateId, appData AppData, lpa *Lpa) error
	UseTestCode()
}

func PostFormString(r *http.Request, name string) string {
	return strings.TrimSpace(r.PostFormValue(name))
}

func PostFormReferenceNumber(r *http.Request, name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(name), " ", ""), "-", "")
}

type Handler func(data AppData, w http.ResponseWriter, r *http.Request) error
