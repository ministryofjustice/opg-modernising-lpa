// Package appcontext provides functionality to pass data in contexts through
// the lifetime of a web request.
package appcontext

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

type Data struct {
	ActorType         actor.Type
	AttorneyUID       actoruid.UID
	CanGoBack         bool
	CookieConsentSet  bool
	CsrfToken         string
	HasLpas           bool
	HideLoginNav      bool
	Lang              localize.Lang
	LoginSessionEmail string
	Localizer         localize.Localizer
	LpaID             string
	Page              string
	Path              string
	Query             url.Values
	SessionID         string
	SupporterData     *SupporterData
}

type SupporterData struct {
	LpaType              lpadata.LpaType
	DonorFullName        string
	OrganisationName     string
	IsManageOrganisation bool
	Permission           supporterdata.Permission
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
