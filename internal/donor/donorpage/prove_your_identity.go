package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type proveYourIdentity struct {
	App                  appcontext.Data
	Errors               validation.List
	LowConfidenceEnabled bool
}

func ProveYourIdentity(tmpl template.Template, lowConfidenceEnabled bool) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		return tmpl(w, &proveYourIdentity{
			App:                  appData,
			LowConfidenceEnabled: lowConfidenceEnabled,
		})
	}
}
