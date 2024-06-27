package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

//go:generate enumerator -type NoVoucherDecision -linecomment -empty
type NoVoucherDecision uint8

const (
	ProveOwnID       NoVoucherDecision = iota + 1 // prove-own-id
	SelectNewVoucher                              // select-new-voucher
	WithdrawLPA                                   // withdraw-lpa
	ApplyToCOP                                    // apply-to-cop
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
				Options: NoVoucherDecisionValues,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readWhatYouCanDoNowForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var next page.LpaPath

				switch data.Form.DoNext {
				case ProveOwnID:
					next = page.Paths.TaskList
				case SelectNewVoucher:
					next = page.Paths.EnterVoucher
				case WithdrawLPA:
					next = page.Paths.WithdrawThisLpa
				case ApplyToCOP:
					donor.WantsToApplyToCourtOfProtection = true
					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}

					next = page.Paths.TaskList
				}

				return next.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type whatYouCanDoNowForm struct {
	DoNext  NoVoucherDecision
	Error   error
	Options NoVoucherDecisionOptions
}

func readWhatYouCanDoNowForm(r *http.Request) *whatYouCanDoNowForm {
	doNext, err := ParseNoVoucherDecision(page.PostFormString(r, "do-next"))

	return &whatYouCanDoNowForm{
		DoNext:  doNext,
		Error:   err,
		Options: NoVoucherDecisionValues,
	}
}

func (f *whatYouCanDoNowForm) Validate() validation.List {
	var errors validation.List

	errors.Error("do-next", "whatYouWouldLikeToDo", f.Error,
		validation.Selected())

	return errors
}
