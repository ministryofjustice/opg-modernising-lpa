package voucherpage

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

func IdentityWithOneLoginCallback(oneLoginClient OneLoginClient, sessionStore SessionStore, voucherStore VoucherStore, lpaStoreResolvingService LpaStoreResolvingService, fail vouchFailer) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return fmt.Errorf("error resolving lpa: %w", err)
		}

		if r.FormValue("error") == "access_denied" {
			// TODO: check with team on how we want to communicate this on the page
			return onelogin.ErrAccessDenied
		}

		oneLoginSession, err := sessionStore.OneLogin(r)
		if err != nil {
			return fmt.Errorf("error getting onelogin session: %w", err)
		}

		_, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return fmt.Errorf("error exchanging code: %w", err)
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return fmt.Errorf("error retrieving userinfo: %w", err)
		}

		userData, err := oneLoginClient.ParseIdentityClaim(userInfo)
		if err != nil {
			return fmt.Errorf("error parsing identity claim: %w", err)
		}

		nameMatches := userData.MatchName(provided.FirstNames, provided.LastName)
		provided.IdentityUserData = userData

		if userData.Status.IsConfirmed() {
			provided.FirstNames = userData.FirstNames
			provided.LastName = userData.LastName
		}

		log.Println(nameMatches, provided.NameMatches(lpa).IsNone())
		if nameMatches || provided.NameMatches(lpa).IsNone() {
			provided.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
		}

		if err := voucherStore.Put(r.Context(), provided); err != nil {
			return fmt.Errorf("error voucher put: %w", err)
		}

		if !provided.IdentityUserData.Status.IsConfirmed() {
			if err := fail(r.Context(), provided, lpa); err != nil {
				return fmt.Errorf("error failing vouch: %w", err)
			}

			return page.PathVoucherUnableToConfirmIdentity.RedirectQuery(w, r, appData, url.Values{
				"donorFullName":   {lpa.Donor.FullName()},
				"donorFirstNames": {lpa.Donor.FirstNames},
			})
		}

		if provided.Tasks.ConfirmYourIdentity.IsCompleted() {
			return voucher.PathOneLoginIdentityDetails.Redirect(w, r, appData, appData.LpaID)
		}

		return voucher.PathConfirmAllowedToVouch.Redirect(w, r, appData, appData.LpaID)
	}
}
