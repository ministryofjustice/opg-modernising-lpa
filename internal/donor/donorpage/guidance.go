package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App    page.AppData
	Errors validation.List
	Donor  *actor.DonorProvidedDetails
}

func Guidance(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &guidanceData{
			App:   appData,
			Donor: donor,
		}

		return tmpl(w, data)
	}
}
