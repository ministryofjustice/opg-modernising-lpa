package fixtures

import (
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
)

func Voucher(
	tmpl template.Template,
	sessionStore *sesh.Store,
	shareCodeStore *sharecode.Store,
	donorStore *donor.Store,
	voucherStore *voucher.Store,
) page.Handler {
	progressValues := []string{
		"confirmYourName",
		"verifyDonorDetails",
		"confirmYourIdentity",
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		var (
			voucherSub = r.FormValue("voucherSub")
			shareCode  = r.FormValue("withShareCode")
			redirect   = r.FormValue("redirect")
			progress   = slices.Index(progressValues, r.FormValue("progress"))
		)

		if voucherSub == "" {
			voucherSub = random.String(16)
		}

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: voucherSub, Email: testEmail}); err != nil {
			return err
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: voucherSub})
		}

		var (
			donorSub         = random.String(16)
			donorSessionID   = base64.StdEncoding.EncodeToString([]byte(donorSub))
			voucherSessionID = base64.StdEncoding.EncodeToString([]byte(voucherSub))
		)

		createSession := &appcontext.Session{SessionID: donorSessionID}
		donorDetails, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), createSession))
		if err != nil {
			return err
		}

		var (
			donorCtx   = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donorDetails.LpaID})
			voucherCtx = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: voucherSessionID, LpaID: donorDetails.LpaID})
		)

		donorDetails.SignedAt = time.Now()
		donorDetails.Donor = makeDonor(testEmail)
		donorDetails.LpaUID = makeUID()
		donorDetails.Type = lpadata.LpaTypePropertyAndAffairs
		donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenHasCapacity
		donorDetails.CertificateProvider = makeCertificateProvider()
		donorDetails.Attorneys = donordata.Attorneys{
			Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0])},
		}
		donorDetails.Voucher = donordata.Voucher{
			UID:        actoruid.New(),
			FirstNames: "Vivian",
			LastName:   "Vaughn",
			Email:      testEmail,
			Allowed:    true,
		}

		if err := donorStore.Put(donorCtx, donorDetails); err != nil {
			return err
		}

		hasShareCode := shareCode != ""
		if !hasShareCode {
			shareCode = random.String(12)
		}

		link := sharecodedata.Link{
			LpaKey:      donorDetails.PK,
			LpaOwnerKey: donorDetails.SK,
			ActorUID:    donorDetails.Voucher.UID,
		}

		if err := shareCodeStore.Put(voucherCtx, actor.TypeVoucher, shareCode, link); err != nil {
			return err
		}

		if hasShareCode {
			http.Redirect(w, r, page.PathVoucherStart.Format(), http.StatusFound)
			return nil
		}

		shareLink, err := shareCodeStore.Get(voucherCtx, actor.TypeVoucher, shareCode)
		if err != nil {
			return err
		}

		voucherDetails, err := voucherStore.Create(voucherCtx, shareLink, testEmail)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "confirmYourName") {
			voucherDetails.FirstNames = donorDetails.Voucher.FirstNames
			voucherDetails.LastName = donorDetails.Voucher.LastName
			voucherDetails.Tasks.ConfirmYourName = task.StateCompleted
		}

		if progress >= slices.Index(progressValues, "verifyDonorDetails") {
			voucherDetails.Tasks.VerifyDonorDetails = task.StateCompleted
		}

		if progress >= slices.Index(progressValues, "confirmYourIdentity") {
			voucherDetails.IdentityUserData = identity.UserData{
				Status:     identity.StatusConfirmed,
				FirstNames: voucherDetails.FirstNames,
				LastName:   voucherDetails.LastName,
			}
			voucherDetails.Tasks.ConfirmYourIdentity = task.StateCompleted
		}

		if err := voucherStore.Put(voucherCtx, voucherDetails); err != nil {
			return err
		}

		if redirect == "" {
			redirect = page.PathDashboard.Format()
		} else {
			redirect = "/voucher/" + donorDetails.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
