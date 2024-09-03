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
	App                 appcontext.Data
	Errors              validation.List
	Form                *whatYouCanDoNowForm
	NewVoucherLabel     string
	BannerContent       string
	FailedVouchAttempts int
}

func WhatYouCanDoNow(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whatYouCanDoNowData{
			App: appData,
			Form: &whatYouCanDoNowForm{
				Options:        donordata.NoVoucherDecisionValues,
				CanHaveVoucher: provided.CanHaveVoucher(),
			},
			FailedVouchAttempts: provided.FailedVouchAttempts,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhatYouCanDoNowForm(r, provided)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var next donor.Path

				switch data.Form.DoNext {
				case donordata.ProveOwnID:
					provided.DonorIdentityUserData = identity.UserData{}
					next = donor.PathTaskList
				case donordata.SelectNewVoucher:
					provided.WantVoucher = form.Yes
					next = donor.PathEnterVoucher
				case donordata.WithdrawLPA:
					next = donor.PathWithdrawThisLpa
				case donordata.ApplyToCOP:
					provided.RegisteringWithCourtOfProtection = true
					next = donor.PathWhatHappensNextRegisteringWithCourtOfProtection
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return next.Redirect(w, r, appData, provided)
			}
		}

		switch provided.FailedVouchAttempts {
		case 0:
			data.BannerContent = "youHaveNotChosenAnyoneToVouchForYou"
			data.NewVoucherLabel = "iHaveSomeoneWhoCanVouch"
		case 1:
			data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinue"
			data.NewVoucherLabel = "iHaveSomeoneElseWhoCanVouch"
		default:
			data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinueSecondAttempt"
		}

		return tmpl(w, data)
	}
}

type whatYouCanDoNowForm struct {
	DoNext         donordata.NoVoucherDecision
	Error          error
	Options        donordata.NoVoucherDecisionOptions
	CanHaveVoucher bool
}

func readWhatYouCanDoNowForm(r *http.Request, provided *donordata.Provided) *whatYouCanDoNowForm {
	doNext, err := donordata.ParseNoVoucherDecision(page.PostFormString(r, "do-next"))

	return &whatYouCanDoNowForm{
		DoNext:         doNext,
		Error:          err,
		Options:        donordata.NoVoucherDecisionValues,
		CanHaveVoucher: provided.CanHaveVoucher(),
	}
}

func (f *whatYouCanDoNowForm) Validate() validation.List {
	var errors validation.List

	errors.Error("do-next", "whatYouWouldLikeToDo", f.Error,
		validation.Selected())

	if !f.CanHaveVoucher && f.DoNext.IsSelectNewVoucher() {
		errors.Add("do-next", validation.CustomError{
			Label: "youCannotAskAnotherPersonToVouchForYou",
		})
	}

	return errors
}
