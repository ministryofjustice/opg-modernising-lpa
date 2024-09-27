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
	ProveOwnIDLabel     string
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
				nextPage := handleDoNext(data.Form.DoNext, provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return nextPage.Redirect(w, r, appData, provided)
			}
		}

		switch provided.FailedVouchAttempts {
		case 0:
			data.BannerContent = "youHaveNotChosenAnyoneToVouchForYou"
			data.NewVoucherLabel = "iHaveSomeoneWhoCanVouch"
			data.ProveOwnIDLabel = "iWillReturnToOneLogin"
		case 1:
			data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinue"
			data.NewVoucherLabel = "iHaveSomeoneElseWhoCanVouch"
			data.ProveOwnIDLabel = "iWillGetOrFindID"
		default:
			data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinueSecondAttempt"
			data.ProveOwnIDLabel = "iWillGetOrFindID"
		}

		return tmpl(w, data)
	}
}

func handleDoNext(doNext donordata.NoVoucherDecision, provided *donordata.Provided) (nextPage donor.Path) {
	switch doNext {
	case donordata.ProveOwnID:
		provided.IdentityUserData = identity.UserData{}
		nextPage = donor.PathTaskList
	case donordata.SelectNewVoucher:
		provided.WantVoucher = form.Yes
		nextPage = donor.PathEnterVoucher
	case donordata.WithdrawLPA:
		nextPage = donor.PathWithdrawThisLpa
	case donordata.ApplyToCOP:
		provided.RegisteringWithCourtOfProtection = true
		nextPage = donor.PathWhatHappensNextRegisteringWithCourtOfProtection
	}

	return nextPage
}

type whatYouCanDoNowForm struct {
	DoNext         donordata.NoVoucherDecision
	Options        donordata.NoVoucherDecisionOptions
	CanHaveVoucher bool
}

func readWhatYouCanDoNowForm(r *http.Request, provided *donordata.Provided) *whatYouCanDoNowForm {
	doNext, _ := donordata.ParseNoVoucherDecision(page.PostFormString(r, "do-next"))

	return &whatYouCanDoNowForm{
		DoNext:         doNext,
		Options:        donordata.NoVoucherDecisionValues,
		CanHaveVoucher: provided.CanHaveVoucher(),
	}
}

func (f *whatYouCanDoNowForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("do-next", "whatYouWouldLikeToDo", f.DoNext,
		validation.Selected())

	if !f.CanHaveVoucher && f.DoNext.IsSelectNewVoucher() {
		errors.Add("do-next", validation.CustomError{
			Label: "youCannotAskAnotherPersonToVouchForYou",
		})
	}

	return errors
}
