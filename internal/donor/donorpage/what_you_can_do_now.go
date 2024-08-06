package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatYouCanDoNowData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *whatYouCanDoNowForm
}

func WhatYouCanDoNow(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whatYouCanDoNowData{
			App: appData,
			Form: &whatYouCanDoNowForm{
				Options: donordata.NoVoucherDecisionValues,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readWhatYouCanDoNowForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var next donor.Path

				switch data.Form.DoNext {
				case donordata.ProveOwnID:
					provided.DonorIdentityUserData = identity.UserData{}
					next = page.Paths.TaskList
				case donordata.SelectNewVoucher:
					provided.WantVoucher = form.Yes
					next = page.Paths.EnterVoucher
				case donordata.WithdrawLPA:
					next = page.Paths.WithdrawThisLpa
				case donordata.ApplyToCOP:
					provided.RegisteringWithCourtOfProtection = true
					next = page.Paths.WhatHappensNextRegisteringWithCourtOfProtection
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return next.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type whatYouCanDoNowForm struct {
	DoNext  donordata.NoVoucherDecision
	Error   error
	Options donordata.NoVoucherDecisionOptions
}

func readWhatYouCanDoNowForm(r *http.Request) *whatYouCanDoNowForm {
	doNext, err := donordata.ParseNoVoucherDecision(page.PostFormString(r, "do-next"))

	return &whatYouCanDoNowForm{
		DoNext:  doNext,
		Error:   err,
		Options: donordata.NoVoucherDecisionValues,
	}
}

func (f *whatYouCanDoNowForm) Validate() validation.List {
	var errors validation.List

	errors.Error("do-next", "whatYouWouldLikeToDo", f.Error,
		validation.Selected())

	return errors
}
