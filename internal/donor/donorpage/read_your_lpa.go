package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readYourLpaData struct {
	App       appcontext.Data
	BannerApp appcontext.Data
	Errors    validation.List
	Donor     *donordata.Provided
}

func ReadYourLpa(tmpl template.Template, bundle Bundle) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		bannerLanguage, err := localize.ParseLang(r.FormValue("bannerLanguage"))
		if err != nil || bannerLanguage.Empty() {
			bannerLanguage = appData.Lang
			appData.Lang = provided.Donor.LpaLanguagePreference
			return donor.PathReadYourLpa.RedirectQuery(w, r, appData, provided, url.Values{
				"bannerLanguage": {bannerLanguage.String()},
			})
		}

		bannerAppData := appData
		bannerAppData.Lang = bannerLanguage
		bannerAppData.Localizer = bundle.For(bannerAppData.Lang)

		data := &readYourLpaData{
			App:       appData,
			BannerApp: bannerAppData,
			Donor:     provided,
		}

		return tmpl(w, data)
	}
}
