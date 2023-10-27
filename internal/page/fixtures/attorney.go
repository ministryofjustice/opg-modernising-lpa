package fixtures

import (
	"context"
	"encoding/base64"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type DonorStore interface {
	Create(context.Context) (*page.Lpa, error)
	Put(context.Context, *page.Lpa) error
}

type CertificateProviderStore interface {
	Create(context.Context, string) (*actor.CertificateProviderProvidedDetails, error)
	Put(context.Context, *actor.CertificateProviderProvidedDetails) error
}

type AttorneyStore interface {
	Create(context.Context, string, string, bool, bool) (*actor.AttorneyProvidedDetails, error)
	Put(context.Context, *actor.AttorneyProvidedDetails) error
}

func Attorney(
	tmpl template.Template,
	sessionStore sesh.Store,
	shareCodeSender ShareCodeSender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
) page.Handler {
	progressValues := []string{
		"signedByCertificateProvider",
		"signedByAttorney",
		"signedByAllAttorneys",
		"submitted",
		"withdrawn",
		"registered",
	}

	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			isReplacement      = r.FormValue("is-replacement") == "1"
			isTrustCorporation = r.FormValue("is-trust-corporation") == "1"
			lpaType            = r.FormValue("lpa-type")
			progress           = slices.Index(progressValues, r.FormValue("progress"))
			email              = r.FormValue("email")
			redirect           = r.FormValue("redirect")
		)

		if r.Method != http.MethodPost && redirect == "" {
			return tmpl(w, &fixturesData{App: appData})
		}

		if lpaType == "hw" && isTrustCorporation {
			return tmpl(w, &fixturesData{App: appData, Errors: validation.With("", validation.CustomError{Label: "Can't add a trust corporation to a personal welfare LPA"})})
		}

		var (
			donorSub                     = random.String(16)
			attorneySub                  = random.String(16)
			certificateProviderSub       = random.String(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
			attorneySessionID            = base64.StdEncoding.EncodeToString([]byte(attorneySub))
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: attorneySub, Email: testEmail}); err != nil {
			return err
		}

		lpa, err := donorStore.Create(page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID}))
		if err != nil {
			return err
		}

		var (
			donorCtx               = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: donorSessionID, LpaID: lpa.ID})
			certificateProviderCtx = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: certificateProviderSessionID, LpaID: lpa.ID})
			attorneyCtx            = page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: attorneySessionID, LpaID: lpa.ID})
		)

		lpa.Donor = actor.Donor{
			FirstNames: "Sam",
			LastName:   "Smith",
			Address: place.Address{
				Line1:      "1 RICHMOND PLACE",
				Line2:      "KINGS HEATH",
				Line3:      "WEST MIDLANDS",
				TownOrCity: "BIRMINGHAM",
				Postcode:   "B14 7ED",
			},
			Email:         testEmail,
			DateOfBirth:   date.New("2000", "1", "2"),
			ThinksCanSign: actor.Yes,
			CanSign:       form.Yes,
		}

		lpa.Type = page.LpaTypePropertyFinance
		if lpaType == "hw" && !isTrustCorporation {
			lpa.Type = page.LpaTypeHealthWelfare
		}

		lpa.CertificateProvider = makeCertificateProvider()

		lpa.Attorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(attorneyNames[0])},
			TrustCorporation: makeTrustCorporation("First Choice Trust Corporation Ltd."),
		}
		lpa.ReplacementAttorneys = actor.Attorneys{
			Attorneys:        []actor.Attorney{makeAttorney(replacementAttorneyNames[0])},
			TrustCorporation: makeTrustCorporation("Second Choice Trust Corporation Ltd."),
		}

		if email != "" {
			if isTrustCorporation {
				if isReplacement {
					lpa.ReplacementAttorneys.TrustCorporation.Email = email
				} else {
					lpa.Attorneys.TrustCorporation.Email = email
				}
			}
			if isReplacement {
				lpa.ReplacementAttorneys.Attorneys[0].Email = email
			} else {
				lpa.Attorneys.Attorneys[0].Email = email
			}
		}

		var attorneyID string
		if !isTrustCorporation {
			if isReplacement {
				attorneyID = lpa.ReplacementAttorneys.Attorneys[0].ID
			} else {
				attorneyID = lpa.Attorneys.Attorneys[0].ID
			}
		}

		certificateProvider, err := certificateProviderStore.Create(certificateProviderCtx, donorSessionID)
		if err != nil {
			return err
		}

		attorney, err := attorneyStore.Create(attorneyCtx, donorSessionID, attorneyID, isReplacement, isTrustCorporation)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			lpa.SignedAt = time.Now()
			certificateProvider.Certificate = actor.Certificate{Agreed: lpa.SignedAt.Add(time.Hour)}
		}
		if progress >= slices.Index(progressValues, "signedByAttorney") {
			attorney.Mobile = testMobile
			attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
			attorney.Tasks.ReadTheLpa = actor.TaskCompleted
			attorney.Tasks.SignTheLpa = actor.TaskCompleted

			if isTrustCorporation {
				attorney.WouldLikeSecondSignatory = form.No
				attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
					FirstNames:        "A",
					LastName:          "Sign",
					ProfessionalTitle: "Assistant to the signer",
					LpaSignedAt:       lpa.SignedAt,
					Confirmed:         lpa.SignedAt.Add(2 * time.Hour),
				}}
			} else {
				attorney.LpaSignedAt = lpa.SignedAt
				attorney.Confirmed = lpa.SignedAt.Add(2 * time.Hour)
			}
		}
		if progress >= slices.Index(progressValues, "signedByAllAttorneys") {
			for isReplacement, list := range map[bool]actor.Attorneys{false: lpa.Attorneys, true: lpa.ReplacementAttorneys} {
				for _, a := range list.Attorneys {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: lpa.ID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, a.ID, isReplacement, false)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.LpaSignedAt = lpa.SignedAt
					attorney.Confirmed = lpa.SignedAt.Add(2 * time.Hour)

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}

				if list.TrustCorporation.Name != "" {
					ctx := page.ContextWithSessionData(r.Context(), &page.SessionData{SessionID: random.String(16), LpaID: lpa.ID})

					attorney, err := attorneyStore.Create(ctx, donorSessionID, "", isReplacement, true)
					if err != nil {
						return err
					}

					attorney.Mobile = testMobile
					attorney.Tasks.ConfirmYourDetails = actor.TaskCompleted
					attorney.Tasks.ReadTheLpa = actor.TaskCompleted
					attorney.Tasks.SignTheLpa = actor.TaskCompleted
					attorney.WouldLikeSecondSignatory = form.No
					attorney.AuthorisedSignatories = [2]actor.TrustCorporationSignatory{{
						FirstNames:        "A",
						LastName:          "Sign",
						ProfessionalTitle: "Assistant to the signer",
						LpaSignedAt:       lpa.SignedAt,
						Confirmed:         lpa.SignedAt.Add(2 * time.Hour),
					}}

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}
				}
			}
		}
		if progress >= slices.Index(progressValues, "submitted") {
			lpa.SubmittedAt = time.Now()
		}
		if progress >= slices.Index(progressValues, "withdrawn") {
			lpa.WithdrawnAt = time.Now()
		}
		if progress >= slices.Index(progressValues, "registered") {
			lpa.RegisteredAt = time.Now()
		}

		if err := donorStore.Put(donorCtx, lpa); err != nil {
			return err
		}
		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}
		if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
			return err
		}

		// should only be used in tests as otherwise people can read their emails...
		if r.FormValue("use-test-code") == "1" {
			shareCodeSender.UseTestCode()
		}

		if email != "" {
			shareCodeSender.SendAttorneys(donorCtx, page.AppData{
				SessionID: donorSessionID,
				LpaID:     lpa.ID,
				Localizer: appData.Localizer,
			}, lpa)

			return page.AppData{}.Redirect(w, r, nil, page.Paths.Attorney.Start.Format())
		}

		if redirect == "" {
			redirect = page.Paths.Dashboard.Format()
		} else {
			redirect = "/attorney/" + lpa.ID + redirect
		}

		return page.AppData{}.Redirect(w, r, nil, redirect)
	}
}
