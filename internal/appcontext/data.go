// Package appcontext provides functionality to pass data in contexts through
// the lifetime of a web request.
package appcontext

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

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

type Data struct {
	Page              string
	Path              string
	Query             url.Values
	Localizer         Localizer
	Lang              localize.Lang
	CookieConsentSet  bool
	CanGoBack         bool
	CanToggleWelsh    bool
	SessionID         string
	LpaID             string
	CsrfToken         string
	ActorType         actor.Type
	AttorneyUID       actoruid.UID
	LoginSessionEmail string
	SupporterData     *SupporterData
}

type SupporterData struct {
	LpaType              lpadata.LpaType
	DonorFullName        string
	OrganisationName     string
	IsManageOrganisation bool
	Permission           actor.Permission
	LoggedInSupporterID  string
}

func ContextWithData(ctx context.Context, appData Data) context.Context {
	return context.WithValue(ctx, (*Data)(nil), appData)
}

func DataFromContext(ctx context.Context) Data {
	appData, _ := ctx.Value((*Data)(nil)).(Data)

	return appData
}

func (d Data) Redirect(w http.ResponseWriter, r *http.Request, url string) error {
	http.Redirect(w, r, d.Lang.URL(url), http.StatusFound)
	return nil
}

func (d Data) IsDonor() bool {
	return d.ActorType == actor.TypeDonor
}

func (d Data) IsCertificateProvider() bool {
	return d.ActorType == actor.TypeCertificateProvider
}

func (d Data) IsAttorneyType() bool {
	return d.ActorType == actor.TypeAttorney ||
		d.ActorType == actor.TypeReplacementAttorney ||
		d.ActorType == actor.TypeTrustCorporation ||
		d.ActorType == actor.TypeReplacementTrustCorporation
}

func (d Data) IsReplacementAttorney() bool {
	return d.ActorType == actor.TypeReplacementAttorney || d.ActorType == actor.TypeReplacementTrustCorporation
}

func (d Data) IsTrustCorporation() bool {
	return d.ActorType == actor.TypeTrustCorporation || d.ActorType == actor.TypeReplacementTrustCorporation
}

func (d Data) IsAdmin() bool {
	return d.SupporterData != nil && d.SupporterData.Permission.IsAdmin()
}

func (d Data) EncodeQuery() string {
	query := ""

	if d.Query.Encode() != "" {
		query = "?" + d.Query.Encode()
	}

	return query
}
