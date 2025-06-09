package certificateproviderpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type readTheLpaData struct {
	App       appcontext.Data
	BannerApp appcontext.Data
	Errors    validation.List
	Lpa       *lpadata.Lpa
}

func ReadTheLpa(tmpl template.Template, certificateProviderStore CertificateProviderStore, bundle Bundle) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		bannerLanguage, err := localize.ParseLang(r.FormValue("bannerLanguage"))
		if err != nil || bannerLanguage.Empty() {
			bannerLanguage = appData.Lang
			appData.Lang = lpa.Language
			return certificateprovider.PathReadTheLpa.RedirectQuery(w, r, appData, lpa.LpaID, url.Values{
				"bannerLanguage": {bannerLanguage.String()},
			})
		}

		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ReadTheLpa = task.StateCompleted
			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			return certificateprovider.PathWhatHappensNext.Redirect(w, r, appData, lpa.LpaID)
		}

		bannerAppData := appData
		bannerAppData.Lang = bannerLanguage
		bannerAppData.Localizer = bundle.For(bannerAppData.Lang)

		data := &readTheLpaData{
			App:       appData,
			BannerApp: bannerAppData,
			Lpa:       lpa,
		}

		return tmpl(w, data)
	}
}
