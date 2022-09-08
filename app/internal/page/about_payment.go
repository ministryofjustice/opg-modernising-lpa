package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type aboutPaymentData struct {
	App    AppData
	Errors map[string]string
}

func AboutPayment(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &aboutPaymentData{
			App: appData,
		}

		return tmpl(w, data)
	}
}
