package page

import (
	"context"
	"net/http"

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
}

func (d AppData) Redirect(w http.ResponseWriter, r *http.Request, lpa *Lpa, url string) error {
	if lpa != nil && d.LpaID == "" {
		d.LpaID = lpa.ID
	}

	// as a shortcut for when you don't have an Lpa but know the transition is fine we allow passing nil
	if lpa == nil || lpa.CanGoTo(url) {
		http.Redirect(w, r, d.BuildUrl(url), http.StatusFound)
	} else {
		http.Redirect(w, r, d.BuildUrl(Paths.TaskList), http.StatusFound)
	}

	return nil
}

func (d AppData) BuildUrl(url string) string {
	if d.Lang == localize.Cy {
		return "/" + localize.Cy.String() + d.BuildUrlWithoutLang(url)
	}

	return d.BuildUrlWithoutLang(url)
}

func (d AppData) BuildUrlWithoutLang(url string) string {
	if IsLpaPath(url) {
		return "/lpa/" + d.LpaID + url
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
