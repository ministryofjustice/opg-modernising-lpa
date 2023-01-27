package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldYouLikeToBeContactedData struct {
	App     AppData
	Errors  validation.List
	Contact []string
}

func HowWouldYouLikeToBeContacted(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howWouldYouLikeToBeContactedData{
			App:     appData,
			Contact: lpa.Contact,
		}

		if r.Method == http.MethodPost {
			form := readHowWouldYouLikeToBeContactedForm(r)
			data.Errors = form.Validate()

			if data.Errors.Empty() {
				lpa.Contact = form.Contact
				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.TaskList)
			}
		}

		return tmpl(w, data)
	}
}

type howWouldYouLikeToBeContactedForm struct {
	Contact []string
}

func readHowWouldYouLikeToBeContactedForm(r *http.Request) *howWouldYouLikeToBeContactedForm {
	r.ParseForm()

	return &howWouldYouLikeToBeContactedForm{
		Contact: r.PostForm["contact"],
	}
}

func (f *howWouldYouLikeToBeContactedForm) Validate() validation.List {
	var errors validation.List

	if len(f.Contact) == 0 {
		errors.Add("contact", "selectContact")
	}

	for _, value := range f.Contact {
		if value != "email" && value != "phone" && value != "text message" && value != "post" {
			errors.Add("contact", "selectContact")
			break
		}
	}

	return errors
}
