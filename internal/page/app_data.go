package page

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type AppData struct {
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
	PublicURL         string
}

type SupporterData struct {
	LpaType              actor.LpaType
	DonorFullName        string
	OrganisationName     string
	IsManageOrganisation bool
	Permission           actor.Permission
	LoggedInSupporterID  string
}

func (d AppData) Redirect(w http.ResponseWriter, r *http.Request, url string) error {
	http.Redirect(w, r, d.Lang.URL(url), http.StatusFound)
	return nil
}

func ContextWithAppData(ctx context.Context, appData AppData) context.Context {
	return context.WithValue(ctx, contextKey("appData"), appData)
}

func AppDataFromContext(ctx context.Context) AppData {
	appData, _ := ctx.Value(contextKey("appData")).(AppData)

	return appData
}

func (d AppData) IsDonor() bool {
	return d.ActorType == actor.TypeDonor
}

func (d AppData) IsCertificateProvider() bool {
	return d.ActorType == actor.TypeCertificateProvider
}

func (d AppData) IsAttorneyType() bool {
	return d.ActorType == actor.TypeAttorney ||
		d.ActorType == actor.TypeReplacementAttorney ||
		d.ActorType == actor.TypeTrustCorporation ||
		d.ActorType == actor.TypeReplacementTrustCorporation
}

func (d AppData) IsReplacementAttorney() bool {
	return d.ActorType == actor.TypeReplacementAttorney || d.ActorType == actor.TypeReplacementTrustCorporation
}

func (d AppData) IsTrustCorporation() bool {
	return d.ActorType == actor.TypeTrustCorporation || d.ActorType == actor.TypeReplacementTrustCorporation
}

func (d AppData) IsAdmin() bool {
	return d.SupporterData != nil && d.SupporterData.Permission.IsAdmin()
}

func (d AppData) EncodeQuery() string {
	query := ""

	if d.Query.Encode() != "" {
		query = "?" + d.Query.Encode()
	}

	return query
}
