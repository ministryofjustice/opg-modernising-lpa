package page

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type AppData struct {
	Page                 string
	Path                 string
	Query                string
	Localizer            Localizer
	Lang                 localize.Lang
	CookieConsentSet     bool
	CanGoBack            bool
	SessionID            string
	LpaID                string
	CsrfToken            string
	ActorType            actor.Type
	AttorneyUID          actoruid.UID
	IsSupporter          bool
	OrganisationName     string
	IsManageOrganisation bool
	LoginSessionEmail    string
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

func (d AppData) IsReplacementAttorney() bool {
	return d.ActorType == actor.TypeReplacementAttorney
}

func (d AppData) IsTrustCorporation() bool {
	return d.ActorType == actor.TypeTrustCorporation || d.ActorType == actor.TypeReplacementTrustCorporation
}
