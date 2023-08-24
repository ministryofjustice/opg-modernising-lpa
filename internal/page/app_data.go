package page

import (
	"context"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
)

type AppData struct {
	Page             string
	Path             string
	Query            string
	Localizer        Localizer
	Lang             localize.Lang
	CookieConsentSet bool
	CanGoBack        bool
	SessionID        string
	RumConfig        RumConfig
	StaticHash       string
	Paths            AppPaths
	LpaID            string
	CsrfToken        string
	ActorTypes       actor.Types
	ActorType        actor.Type
	AttorneyID       string
	OneloginURL      string
}

func (d AppData) Redirect(w http.ResponseWriter, r *http.Request, lpa *Lpa, url string) error {
	if fromURL := r.FormValue("from"); fromURL != "" {
		url = fromURL
	}

	if lpa != nil && d.LpaID == "" {
		d.LpaID = lpa.ID
	}

	// as a shortcut for when you don't have an Lpa but know the transition is fine we allow passing nil
	if lpa == nil || lpa.CanGoTo(url) {
		http.Redirect(w, r, d.BuildUrl(url), http.StatusFound)
	} else {
		http.Redirect(w, r, d.BuildUrl(Paths.TaskList.Format(d.LpaID)), http.StatusFound)
	}

	return nil
}

func (d AppData) BuildUrl(url string) string {
	if d.Lang == localize.Cy {
		return "/" + localize.Cy.String() + url
	}

	return url
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
	return (d.ActorType == actor.TypeAttorney || d.ActorType == actor.TypeReplacementAttorney) && d.AttorneyID == ""
}
