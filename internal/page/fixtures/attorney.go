package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type CertificateProviderStore interface {
	Create(ctx context.Context, link accesscodedata.Link, email string) (*certificateproviderdata.Provided, error)
	Put(ctx context.Context, certificateProvider *certificateproviderdata.Provided) error
}

type AttorneyStore interface {
	Create(ctx context.Context, link accesscodedata.Link, email string) (*attorneydata.Provided, error)
	Put(ctx context.Context, attorney *attorneydata.Provided) error
}

type ScheduledStore interface {
	Create(ctx context.Context, rows ...scheduled.Event) error
}

func Attorney(
	tmpl template.Template,
	sessionStore *sesh.Store,
	accessCodeSender *accesscode.Sender,
	donorStore *donor.Store,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
	organisationStore OrganisationStore,
	memberStore *supporter.MemberStore,
	accessCodeStore *accesscode.Store,
	dynamoClient DynamoClient,
	scheduledStore ScheduledStore,
) page.Handler {
	progressValues := []string{
		"signedByCertificateProvider",
		"confirmYourDetails",
		"readTheLPA",
		"signedByAttorney",
		"signedByAllAttorneys",
		"withdrawn",
		"registered",
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		var (
			lpaType     = r.FormValue("lpa-type")
			lpaLanguage = r.FormValue("lpa-language")
			email       = r.FormValue("email")
			redirect    = r.FormValue("redirect")
			attorneySub = r.FormValue("attorneySub")
			accessCode  = r.FormValue("withAccessCode")

			progress = slices.Index(progressValues, r.FormValue("progress"))

			options            = r.Form["options"]
			useRealUID         = slices.Contains(options, "uid")
			isReplacement      = slices.Contains(options, "is-replacement")
			isTrustCorporation = slices.Contains(options, "is-trust-corporation")
			isSupported        = slices.Contains(options, "is-supported")
			isPaperDonor       = slices.Contains(options, "is-paper-donor")
			isPaperAttorney    = slices.Contains(options, "is-paper-attorney")
			hasPhoneNumber     = slices.Contains(options, "has-phone-number")
			withTestSchedule   = slices.Contains(options, "with-test-schedule")
		)

		if attorneySub == "" {
			attorneySub = random.AlphaNumeric(16)
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: attorneySub})
		}

		if lpaType == "personal-welfare" && isTrustCorporation {
			return tmpl(w, &fixturesData{App: appData, Errors: validation.With("", validation.CustomError{Label: "Can't add a trust corporation to a personal welfare LPA"})})
		}

		encodedSub := encodeSub(attorneySub)

		var (
			donorSub                     = random.AlphaNumeric(16)
			certificateProviderSub       = random.AlphaNumeric(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(certificateProviderSub))
			attorneySessionID            = base64.StdEncoding.EncodeToString([]byte(mockGOLSubPrefix + encodedSub))
		)

		if err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: mockGOLSubPrefix + encodedSub, Email: testEmail, HasLPAs: true}); err != nil {
			return err
		}

		var donorDetails *donordata.Provided
		if isPaperDonor {
			lpaID := random.UUID()

			donorDetails = &donordata.Provided{
				PK:        dynamo.LpaKey(lpaID),
				SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:     lpaID,
				LpaUID:    makeUID(),
				CreatedAt: time.Now(),
				Version:   1,
			}

			transaction := dynamo.NewTransaction().
				Create(donorDetails).
				Create(dynamo.Keys{PK: dynamo.UIDKey(donorDetails.LpaUID), SK: dynamo.MetadataKey("")}).
				Create(dynamo.Keys{PK: donorDetails.PK, SK: dynamo.ReservedKey(dynamo.DonorKey)})

			if err := dynamoClient.WriteTransaction(r.Context(), transaction); err != nil {
				return fmt.Errorf("could not write paper donor %s: %w", lpaID, err)
			}
		} else {
			createFn := donorStore.Create
			createSession := &appcontext.Session{SessionID: donorSessionID}
			if isSupported {
				createFn = organisationStore.CreateLPA

				supporterCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, Email: testEmail})

				member, err := memberStore.Create(supporterCtx, random.AlphaNumeric(12), random.AlphaNumeric(12))
				if err != nil {
					return err
				}

				org, err := organisationStore.Create(supporterCtx, member, random.AlphaNumeric(12))
				if err != nil {
					return err
				}

				createSession.OrganisationID = org.ID
			}
			var err error
			donorDetails, err = createFn(appcontext.ContextWithSession(r.Context(), createSession))
			if err != nil {
				return err
			}

			if isSupported {
				if err := donorStore.Link(appcontext.ContextWithSession(r.Context(), createSession), accesscodedata.Link{
					LpaKey:      donorDetails.PK,
					LpaOwnerKey: donorDetails.SK,
					LpaUID:      donorDetails.LpaUID,
				}, donorDetails.Donor.Email); err != nil {
					return err
				}
			}
		}

		var (
			donorCtx               = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donorDetails.LpaID})
			certificateProviderCtx = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: certificateProviderSessionID, LpaID: donorDetails.LpaID})
			attorneyCtx            = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: attorneySessionID, LpaID: donorDetails.LpaID})
		)

		donorDetails.SignedAt = testNow
		donorDetails.WitnessedByCertificateProviderAt = testNow
		donorDetails.Donor = makeDonor(testEmail, "Sam", "Smith")
		donorDetails.Donor.LpaLanguagePreference, _ = localize.ParseLang(lpaLanguage)
		if donorDetails.Donor.LpaLanguagePreference.Empty() {
			donorDetails.Donor.LpaLanguagePreference = localize.En
		}

		if lpaType == "personal-welfare" && !isTrustCorporation {
			donorDetails.Type = lpadata.LpaTypePersonalWelfare
			donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenCapacityLost
			donorDetails.LifeSustainingTreatmentOption = lpadata.LifeSustainingTreatmentOptionA
		} else {
			donorDetails.Type = lpadata.LpaTypePropertyAndAffairs
			donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenHasCapacity
		}

		donorDetails.Restrictions = makeRestriction(donorDetails)

		if useRealUID {
			err := donorStore.Put(donorCtx, donorDetails)
			if err != nil {
				return err
			}

			if err := eventClient.SendUidRequested(r.Context(), event.UidRequested{
				LpaID:          donorDetails.LpaID,
				DonorSessionID: donorSessionID,
				Type:           donorDetails.Type.String(),
				Donor: uid.DonorDetails{
					Name:     donorDetails.Donor.FullName(),
					Dob:      donorDetails.Donor.DateOfBirth,
					Postcode: donorDetails.Donor.Address.Postcode,
				},
			}); err != nil {
				return err
			}
			donorDetails.LpaUID = waitForRealUID(15, donorStore, donorCtx)

			donorDetails, err = donorStore.Get(donorCtx)
			if err != nil {
				return err
			}
		} else if donorDetails.LpaUID == "" {
			donorDetails.LpaUID = makeUID()
		}

		donorDetails.CertificateProvider = makeCertificateProvider()

		donorDetails.Attorneys = donordata.Attorneys{
			Attorneys:        []donordata.Attorney{makeAttorney(attorneyNames[0])},
			TrustCorporation: makeTrustCorporation("First Choice Trust Corporation Ltd."),
		}
		donorDetails.ReplacementAttorneys = donordata.Attorneys{
			Attorneys:        []donordata.Attorney{makeAttorney(replacementAttorneyNames[0])},
			TrustCorporation: makeTrustCorporation("Second Choice Trust Corporation Ltd."),
		}

		if email != "" {
			if isTrustCorporation && isReplacement {
				donorDetails.ReplacementAttorneys.TrustCorporation.Email = email
			} else if isTrustCorporation {
				donorDetails.Attorneys.TrustCorporation.Email = email
			} else if isReplacement {
				donorDetails.ReplacementAttorneys.Attorneys[0].Email = email
			} else {
				donorDetails.Attorneys.Attorneys[0].Email = email
			}
		}

		var attorneyUID actoruid.UID
		if isTrustCorporation && isReplacement {
			attorneyUID = donorDetails.ReplacementAttorneys.TrustCorporation.UID
		} else if isTrustCorporation {
			attorneyUID = donorDetails.Attorneys.TrustCorporation.UID
		} else if isReplacement {
			attorneyUID = donorDetails.ReplacementAttorneys.Attorneys[0].UID
		} else {
			attorneyUID = donorDetails.Attorneys.Attorneys[0].UID
		}

		donorDetails.AttorneyDecisions = donordata.AttorneyDecisions{How: lpadata.JointlyAndSeverally}
		donorDetails.ReplacementAttorneyDecisions = donordata.AttorneyDecisions{How: lpadata.JointlyAndSeverally}
		donorDetails.HowShouldReplacementAttorneysStepIn = lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct
		donorDetails.Tasks.PayForLpa = task.PaymentStateCompleted

		certificateProvider, err := createCertificateProvider(certificateProviderCtx, accessCodeStore, certificateProviderStore, donorDetails)
		if err != nil {
			return err
		}

		certificateProvider.ContactLanguagePreference = localize.En
		certificateProvider.SignedAt = testNow

		attorney, err := createAttorney(
			attorneyCtx,
			accessCodeStore,
			attorneyStore,
			donorDetails,
			attorneyUID,
			isReplacement,
			isTrustCorporation,
			email,
		)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "signedByCertificateProvider") {
			donorDetails.SignedAt = time.Now()
			certificateProvider.SignedAt = donorDetails.SignedAt.Add(time.Hour)

			if withTestSchedule {
				if err = scheduledStore.Create(attorneyCtx, scheduled.Event{
					At:                time.Now(),
					Action:            scheduleddata.ActionRemindAttorneyToComplete,
					TargetLpaKey:      donorDetails.PK,
					TargetLpaOwnerKey: donorDetails.SK,
					LpaUID:            donorDetails.LpaUID,
				}); err != nil {
					return fmt.Errorf("error scheduling attorneys prompt: %w", err)
				}

				donorDetails.AttorneysInvitedAt = time.Now().AddDate(0, -4, 0)
				donorDetails.IdentityUserData.CheckedAt = time.Now()
				donorDetails.SignedAt = time.Now().AddDate(-1, -10, 0)
			}
		}

		if !isPaperAttorney {
			if progress >= slices.Index(progressValues, "confirmYourDetails") {
				attorney.Phone = testMobile
				attorney.ContactLanguagePreference = localize.En
				attorney.Tasks.ConfirmYourDetails = task.StateCompleted
			}

			if progress >= slices.Index(progressValues, "readTheLPA") {
				attorney.Tasks.ReadTheLpa = task.StateCompleted
			}

			if progress >= slices.Index(progressValues, "signedByAttorney") {
				attorney.Tasks.SignTheLpa = task.StateCompleted

				if isTrustCorporation {
					attorney.WouldLikeSecondSignatory = form.No
					attorney.AuthorisedSignatories = [2]attorneydata.TrustCorporationSignatory{{
						FirstNames:        "A",
						LastName:          "Sign",
						ProfessionalTitle: "Assistant to the signer",
						SignedAt:          donorDetails.SignedAt.Add(2 * time.Hour),
					}}
				} else {
					attorney.SignedAt = donorDetails.SignedAt.Add(2 * time.Hour)
				}
			}
		}

		var signings []*attorneydata.Provided
		if progress >= slices.Index(progressValues, "signedByAllAttorneys") {
			for isReplacement, list := range map[bool]donordata.Attorneys{false: donorDetails.Attorneys, true: donorDetails.ReplacementAttorneys} {
				for _, a := range list.Attorneys {
					ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.AlphaNumeric(16), LpaID: donorDetails.LpaID})

					attorney, err := createAttorney(
						ctx,
						accessCodeStore,
						attorneyStore,
						donorDetails,
						a.UID,
						isReplacement,
						false,
						a.Email,
					)
					if err != nil {
						return err
					}

					attorney.Phone = testMobile
					attorney.ContactLanguagePreference = localize.En
					attorney.Tasks.ConfirmYourDetails = task.StateCompleted
					attorney.Tasks.ReadTheLpa = task.StateCompleted
					attorney.Tasks.SignTheLpa = task.StateCompleted
					attorney.SignedAt = donorDetails.SignedAt.Add(2 * time.Hour)

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}

					signings = append(signings, attorney)
				}

				if list.TrustCorporation.Name != "" {
					ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: random.AlphaNumeric(16), LpaID: donorDetails.LpaID})

					attorney, err := createAttorney(
						ctx,
						accessCodeStore,
						attorneyStore,
						donorDetails,
						list.TrustCorporation.UID,
						isReplacement,
						true,
						list.TrustCorporation.Email,
					)
					if err != nil {
						return err
					}

					attorney.Phone = testMobile
					attorney.ContactLanguagePreference = localize.En
					attorney.Tasks.ConfirmYourDetails = task.StateCompleted
					attorney.Tasks.ReadTheLpa = task.StateCompleted
					attorney.Tasks.SignTheLpa = task.StateCompleted
					attorney.WouldLikeSecondSignatory = form.No
					attorney.AuthorisedSignatories = [2]attorneydata.TrustCorporationSignatory{{
						FirstNames:        "A",
						LastName:          "Sign",
						ProfessionalTitle: "Assistant to the signer",
						SignedAt:          donorDetails.SignedAt.Add(2 * time.Hour),
					}}

					if err := attorneyStore.Put(ctx, attorney); err != nil {
						return err
					}

					signings = append(signings, attorney)
				}
			}
		}

		if progress == slices.Index(progressValues, "withdrawn") {
			donorDetails.WithdrawnAt = time.Now()
		}

		registered := false
		if progress >= slices.Index(progressValues, "registered") {
			registered = true
		}

		if !isPaperDonor {
			if err := donorStore.Put(donorCtx, donorDetails); err != nil {
				return err
			}
		}

		if donorDetails.LpaUID != "" {
			body := lpastore.CreateLpaFromDonorProvided(donorDetails)

			if hasPhoneNumber {
				if isTrustCorporation && isReplacement {
					body.TrustCorporations[1].Mobile = testMobile
				} else if isTrustCorporation {
					body.TrustCorporations[0].Mobile = testMobile
				} else if isReplacement {
					body.Attorneys[1].Mobile = testMobile
				} else {
					body.Attorneys[0].Mobile = testMobile
				}
			}

			if isPaperAttorney {
				body.Attorneys[0].Channel = lpadata.ChannelPaper
				body.Attorneys[0].Address = donorDetails.Donor.Address
				body.Attorneys[0].Email = ""

				if progress >= slices.Index(progressValues, "signedByAttorney") {
					body.Attorneys[0].SignedAt = &testNow
				}
			}

			if err := lpaStoreClient.SendLpa(donorCtx, donorDetails.LpaUID, body); err != nil {
				return fmt.Errorf("problem sending lpa: %w", err)
			}
		}

		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}
		if err := attorneyStore.Put(attorneyCtx, attorney); err != nil {
			return err
		}

		if donorDetails.LpaUID != "" {
			lpa, err := lpaStoreClient.Lpa(donorCtx, donorDetails.LpaUID)
			if err != nil {
				return fmt.Errorf("problem getting lpa: %w", err)
			}

			if err := lpaStoreClient.SendCertificateProvider(donorCtx, certificateProvider, lpa); err != nil {
				return fmt.Errorf("problem sending certificate provider: %w", err)
			}

			for _, attorney := range signings {
				if err := lpaStoreClient.SendAttorney(donorCtx, lpa, attorney); err != nil {
					return fmt.Errorf("problem sending attorney: %w", err)
				}
			}

			if progress >= slices.Index(progressValues, "signedByAllAttorneys") {
				if err := lpaStoreClient.SendStatutoryWaitingPeriod(donorCtx, donorDetails.LpaUID); err != nil {
					return fmt.Errorf("problem sending statutory waiting period: %w", err)
				}
			}

			if registered {
				if err := lpaStoreClient.SendRegister(donorCtx, donorDetails.LpaUID); err != nil {
					return fmt.Errorf("problem sending register: %w", err)
				}
			}
		}

		// should only be used in tests as otherwise people can read their emails...
		if accessCode != "" {
			accessCodeSender.UseTestCode(accessCode)
		}

		if email != "" {
			lpa, err := lpaStoreClient.Lpa(r.Context(), donorDetails.LpaUID)
			if err != nil {
				return err
			}

			lpa.LpaKey = donorDetails.PK
			lpa.LpaOwnerKey = donorDetails.SK

			if isTrustCorporation && isReplacement {
				lpa.Attorneys = lpadata.Attorneys{}
				lpa.ReplacementAttorneys = lpadata.Attorneys{
					TrustCorporation: lpa.ReplacementAttorneys.TrustCorporation,
				}
			} else if isTrustCorporation {
				lpa.Attorneys = lpadata.Attorneys{
					TrustCorporation: lpa.Attorneys.TrustCorporation,
				}
				lpa.ReplacementAttorneys = lpadata.Attorneys{}
			} else if isReplacement {
				lpa.Attorneys = lpadata.Attorneys{}
				lpa.ReplacementAttorneys = lpadata.Attorneys{
					Attorneys: lpa.ReplacementAttorneys.Attorneys,
				}
			} else {
				lpa.Attorneys = lpadata.Attorneys{
					Attorneys: lpa.Attorneys.Attorneys,
				}
				lpa.ReplacementAttorneys = lpadata.Attorneys{}
			}

			accessCodeSender.SendAttorneys(donorCtx, appcontext.Data{
				SessionID: donorSessionID,
				LpaID:     donorDetails.LpaID,
				Localizer: appData.Localizer,
			}, lpa)

			http.Redirect(w, r, page.PathAttorneyStart.Format(), http.StatusFound)
			return nil
		}

		if redirect == "" {
			redirect = page.PathDashboard.Format()
		} else {
			redirect = "/attorney/" + donorDetails.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
