package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type youHaveDecidedNotToBeACertificateProviderData struct {
	App           page.AppData
	Errors        validation.List
	DonorFullName string
}

func YouHaveDecidedNotToBeACertificateProvider(tmpl template.Template) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		return tmpl(w, youHaveDecidedNotToBeACertificateProviderData{
			App:           appData,
			DonorFullName: r.URL.Query().Get("donorFullName"),
		})
	}
}
