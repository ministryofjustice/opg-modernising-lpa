package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whoIsTheLpaForData struct {
	App    page.AppData
	Errors validation.List
	WhoFor string
}

func WhoIsTheLpaFor(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &whoIsTheLpaForData{
			App:    appData,
			WhoFor: lpa.WhoFor,
		}

		if r.Method == http.MethodPost {
			f := readWhoIsTheLpaForForm(r)
			data.Errors = f.Validate()

			if data.Errors.None() {
				lpa.WhoFor = f.WhoFor
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.LpaType)
			}
		}

		return tmpl(w, data)
	}
}

type whoIsTheLpaForForm struct {
	WhoFor string
}

func readWhoIsTheLpaForForm(r *http.Request) *whoIsTheLpaForForm {
	return &whoIsTheLpaForForm{
		WhoFor: page.PostFormString(r, "who-for"),
	}
}

func (f *whoIsTheLpaForForm) Validate() validation.List {
	var errors validation.List

	errors.String("who-for", "whoTheLpaIsFor", f.WhoFor,
		validation.Select("me", "someone-else"))

	return errors
}
