package fixtures

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
)

func Voucher(tmpl template.Template, shareCodeStore *sharecode.Store, donorStore *donor.Store) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		var (
			voucherSub = r.FormValue("voucherSub")
			shareCode  = r.FormValue("withShareCode")
		)

		if voucherSub == "" {
			voucherSub = random.String(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: voucherSub})
		}

		var (
			donorSub       = random.String(16)
			donorSessionID = base64.StdEncoding.EncodeToString([]byte(donorSub))
		)

		createSession := &appcontext.Session{SessionID: donorSessionID}
		donorDetails, err := donorStore.Create(appcontext.ContextWithSession(r.Context(), createSession))
		if err != nil {
			return err
		}

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
			UID: actoruid.New(),
		}

		if shareCode != "" {
			if err := shareCodeStore.Put(r.Context(), actor.TypeVoucher, shareCode, sharecodedata.Link{
				LpaKey:      donorDetails.PK,
				LpaOwnerKey: donorDetails.SK,
				ActorUID:    donorDetails.Voucher.UID,
			}); err != nil {
				return err
			}
		}

		http.Redirect(w, r, page.PathVoucherStart.Format(), http.StatusFound)
		return nil
	}
}
