package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

type addAnLPAData struct {
	App  appcontext.Data
	Form *newforms.EnumForm[actor.Type, actor.TypeOptions, *actor.Type]
}

func AddAnLPA(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := addAnLPAData{
			App:  appData,
			Form: newforms.NewEnumForm[actor.Type](appData.Localizer.T("youMustSelectATypeOfAccessCodeToContinue"), actor.TypeValues),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			switch data.Form.Enum.Value {
			case actor.TypeDonor:
				http.Redirect(w, r, PathEnterAccessCode.Format(), http.StatusFound)
			case actor.TypeCertificateProvider:
				http.Redirect(w, r, PathCertificateProviderEnterAccessCode.Format(), http.StatusFound)
			case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
				http.Redirect(w, r, PathAttorneyEnterAccessCode.Format(), http.StatusFound)
			case actor.TypeVoucher:
				http.Redirect(w, r, PathVoucherEnterAccessCode.Format(), http.StatusFound)
			default:
				http.Redirect(w, r, PathDashboard.Format(), http.StatusFound)
			}
			return nil
		}

		return tmpl(w, data)
	}
}
