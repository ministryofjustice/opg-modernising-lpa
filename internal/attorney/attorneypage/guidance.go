package attorneypage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

type guidanceData struct {
	App appcontext.Data
	Lpa *lpadata.Lpa
}

func Guidance(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, _ *attorneydata.Provided, lpa *lpadata.Lpa) error {
		data := &guidanceData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
