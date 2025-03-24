package donorpage

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
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
	VouchStatusContent    string
	Donor                 *donordata.Provided
}

func WhatYouCanDoNow(tmpl template.Template, donorStore DonorStore, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whatYouCanDoNowData{
			App: appData,
			Form: &whatYouCanDoNowForm{
				Options:        donordata.NoVoucherDecisionValues,
				CanHaveVoucher: provided.CanHaveVoucher(),
			},
			Donor: provided,
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

		voucher, err := voucherStore.GetAny(r.Context())
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return fmt.Errorf("error getting voucher: %w", err)
		}

		if !provided.Voucher.Allowed && voucher == nil {
			switch provided.VouchAttempts {
			case 0:
				data.BannerContent = "youHaveNotChosenAnyoneToVouchForYou"
				data.NewVoucherLabel = "iHaveSomeoneWhoCanVouch"
				data.ProveOwnIdentityLabel = "iWillReturnToOneLogin"
			case 1:
				data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinue"
				data.VouchStatusContent = "tryVouchingAgainContent"
				data.NewVoucherLabel = "iHaveSomeoneElseWhoCanVouch"
				data.ProveOwnIdentityLabel = "iWillGetOrFindID"
			default:
				data.BannerContent = "thePersonYouAskedToVouchHasBeenUnableToContinueSecondAttempt"
				data.ProveOwnIdentityLabel = "iWillGetOrFindID"
			}
		} else if voucher != nil {
			data.BannerContent = "voucherHasNotStartedTheProcess"
			data.VouchStatusContent = "voucherHasNotStartedTheProcessContent"
			data.NewVoucherLabel = "iHaveSomeoneElseWhoCanVouch"
			data.ProveOwnIdentityLabel = "iWillGetOrFindID"

			if provided.VouchAttempts == 1 && voucher.Tasks.VerifyDonorDetails.IsCompleted() {
				data.BannerContent = "voucherHasNotCompletedTheProcess"
				data.VouchStatusContent = "voucherHasNotCompletedTheProcessContent"
			}

			if provided.VouchAttempts > 1 && voucher.Tasks.VerifyDonorDetails.IsCompleted() {
				data.BannerContent = "voucherHasNotCompletedTheProcessSuggestContactVoucher"
				data.VouchStatusContent = ""
			}
		}

		return tmpl(w, data)
	}
}

func handleDoNext(doNext donordata.NoVoucherDecision, provided *donordata.Provided) donor.Path {
	switch doNext {
	case donordata.ProveOwnIdentity:
		provided.WantVoucher = form.No
		return donor.PathConfirmYourIdentity
	case donordata.SelectNewVoucher:
		provided.WantVoucher = form.Yes
		return donor.PathEnterVoucher
	case donordata.WithdrawLPA:
		provided.WantVoucher = form.No
		return donor.PathWithdrawThisLpa
	case donordata.ApplyToCOP:
		provided.WantVoucher = form.No
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
