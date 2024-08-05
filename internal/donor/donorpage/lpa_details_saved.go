package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type LpaDetailsSavedData struct {
	App          appcontext.Data
	Donor        *donordata.Provided
	IsFirstCheck bool
	Errors       validation.List
}

func LpaDetailsSaved(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		return tmpl(w, LpaDetailsSavedData{
			App:          appData,
			IsFirstCheck: r.URL.Query().Has("firstCheck"),
			Donor:        donor,
		})
	}
}
