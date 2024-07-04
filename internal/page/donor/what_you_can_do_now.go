package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatYouCanDoNowData struct {
	App    page.AppData
	Errors validation.List
	Form   *whatYouCanDoNowForm
}

func WhatYouCanDoNow(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &whatYouCanDoNowData{
			App: appData,
			Form: &whatYouCanDoNowForm{
				Options: actor.NoVoucherDecisionValues,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readWhatYouCanDoNowForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var next page.LpaPath

				switch data.Form.DoNext {
				case actor.ProveOwnID:
					donor.DonorIdentityUserData = identity.UserData{}
					next = page.Paths.TaskList
				case actor.SelectNewVoucher:
					donor.WantVoucher = form.Yes
					next = page.Paths.EnterVoucher
				case actor.WithdrawLPA:
					next = page.Paths.WithdrawThisLpa
				case actor.ApplyToCOP:
					donor.RegisteringWithCourtOfProtection = true
					next = page.Paths.WhatHappensNextRegisteringWithCourtOfProtection
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return next.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type whatYouCanDoNowForm struct {
	DoNext  actor.NoVoucherDecision
	Error   error
	Options actor.NoVoucherDecisionOptions
}

func readWhatYouCanDoNowForm(r *http.Request) *whatYouCanDoNowForm {
	doNext, err := actor.ParseNoVoucherDecision(page.PostFormString(r, "do-next"))

	return &whatYouCanDoNowForm{
		DoNext:  doNext,
		Error:   err,
		Options: actor.NoVoucherDecisionValues,
	}
}

func (f *whatYouCanDoNowForm) Validate() validation.List {
	var errors validation.List

	errors.Error("do-next", "whatYouWouldLikeToDo", f.Error,
		validation.Selected())

	return errors
}
