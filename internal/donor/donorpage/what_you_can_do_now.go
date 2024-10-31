package donorpage

import (
	"net/http"
	"net/url"

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
	App                   appcontext.Data
	Errors                validation.List
	Form                  *whatYouCanDoNowForm
	ProveOwnIdentityLabel string
	NewVoucherLabel       string
	BannerContent         string
	FailedVouchAttempts   int
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
				var nextPage string
				if provided.Voucher.Allowed {
					nextPage = donor.PathAreYouSureYouNoLongerNeedVoucher.FormatQuery(provided.LpaID, url.Values{
						"choice": {data.Form.DoNext.String()},
					})
				} else {
					nextPage = handleDoNext(data.Form.DoNext, provided).Format(provided.LpaID)
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return appData.Redirect(w, r, nextPage)
			}
		}

		if !provided.Voucher.Allowed {
			switch provided.FailedVouchAttempts {
			case 0:
				data.BannerContent = "youHaveNotChosenAnyoneToVouchForYou"
				data.NewVoucherLabel = "iHaveSomeoneWhoCanVouch"
				data.ProveOwnIdentityLabel = "iWillReturnToOneLogin"
			case 1:
				data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinue"
				data.NewVoucherLabel = "iHaveSomeoneElseWhoCanVouch"
				data.ProveOwnIdentityLabel = "iWillGetOrFindID"
			default:
				data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinueSecondAttempt"
				data.ProveOwnIdentityLabel = "iWillGetOrFindID"
			}
		} else {
			data.NewVoucherLabel = "iHaveSomeoneElseWhoCanVouch"
			data.ProveOwnIdentityLabel = "iWillGetOrFindID"
		}

		return tmpl(w, data)
	}
}

func handleDoNext(doNext donordata.NoVoucherDecision, provided *donordata.Provided) donor.Path {
	switch doNext {
	case donordata.ProveOwnIdentity:
		provided.IdentityUserData = identity.UserData{}
		return donor.PathTaskList
	case donordata.SelectNewVoucher:
		provided.WantVoucher = form.Yes
		return donor.PathEnterVoucher
	case donordata.WithdrawLPA:
		return donor.PathWithdrawThisLpa
	case donordata.ApplyToCOP:
		provided.RegisteringWithCourtOfProtection = true
		return donor.PathWhatHappensNextRegisteringWithCourtOfProtection
	}

	panic("doNext invalid")
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
