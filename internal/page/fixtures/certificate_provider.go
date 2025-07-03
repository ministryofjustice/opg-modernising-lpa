package fixtures

import (
	"cmp"
	"encoding/base64"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

func CertificateProvider(
	tmpl template.Template,
	sessionStore *sesh.Store,
	accessCodeSender *accesscode.Sender,
	donorStore DonorStore,
	certificateProviderStore CertificateProviderStore,
	eventClient *event.Client,
	lpaStoreClient *lpastore.Client,
	dynamoClient DynamoClient,
	organisationStore OrganisationStore,
	memberStore *supporter.MemberStore,
	accessCodeStore *accesscode.Store,
) page.Handler {
	progressValues := []string{
		"paid",
		"signedByDonor",
		"confirmYourDetails",
		"confirmYourIdentity",
		"provideYourCertificate",
	}

	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		acceptCookiesConsent(w)

		var (
			email                  = r.FormValue("email")
			phone                  = r.FormValue("phone")
			certificateProviderSub = cmp.Or(r.FormValue("certificateProviderSub"), random.AlphaNumeric(16))
			donorEmail             = cmp.Or(r.FormValue("donorEmail"), testEmail)
			lpaType                = r.FormValue("lpa-type")
			lpaLanguage, _         = localize.ParseLang(r.FormValue("lpa-language"))

			options                           = r.Form["options"]
			useRealUID                        = slices.Contains(options, "uid")
			fromStartPage                     = slices.Contains(options, "from-start-page")
			asProfessionalCertificateProvider = slices.Contains(options, "is-professional")
			isSupported                       = slices.Contains(options, "is-supported")
			isPaperDonor                      = slices.Contains(options, "is-paper-donor")

			progress = slices.Index(progressValues, r.FormValue("progress"))

			redirect                   = r.FormValue("redirect")
			accessCode                 = r.FormValue("withAccessCode")
			idStatus                   = r.FormValue("idStatus")
			certificateProviderChannel = r.FormValue("certificateProviderChannel")
		)

		if lpaLanguage.Empty() {
			lpaLanguage = localize.En
		}

		if fromStartPage {
			redirect = "/certificate-provider-start"
		}

		if phone == "not-provided" {
			phone = ""
		} else if phone == "" {
			phone = testMobile
		}

		if r.Method != http.MethodPost && !r.URL.Query().Has("redirect") {
			return tmpl(w, &fixturesData{App: appData, Sub: certificateProviderSub, DonorEmail: donorEmail})
		}

		encodedSub := encodeSub(certificateProviderSub)

		var (
			donorSub                     = random.AlphaNumeric(16)
			donorSessionID               = base64.StdEncoding.EncodeToString([]byte(donorSub))
			certificateProviderSessionID = base64.StdEncoding.EncodeToString([]byte(mockGOLSubPrefix + encodedSub))
		)

		err := sessionStore.SetLogin(r, w, &sesh.LoginSession{Sub: mockGOLSubPrefix + encodedSub, Email: testEmail, HasLPAs: true})
		if err != nil {
			return err
		}

		channel := lpadata.ChannelOnline
		if certificateProviderChannel != "" {
			channel, err = lpadata.ParseChannel(certificateProviderChannel)
			if err != nil {
				return errors.New("invalid format for certificateProviderChannel")
			}
		}

		var donorDetails *donordata.Provided

		if isPaperDonor {
			lpaID := random.UUID()
			donorDetails = &donordata.Provided{
				PK:                               dynamo.LpaKey(lpaID),
				SK:                               dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:                            lpaID,
				LpaUID:                           makeUID(),
				CreatedAt:                        time.Now(),
				Version:                          1,
				HasSentApplicationUpdatedEvent:   true,
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
			}

			transaction := dynamo.NewTransaction().
				Create(donorDetails).
				Create(dynamo.Keys{PK: dynamo.UIDKey(donorDetails.LpaUID), SK: dynamo.MetadataKey("")}).
				Create(dynamo.Keys{PK: donorDetails.PK, SK: dynamo.ReservedKey(dynamo.DonorKey)})

			if err := dynamoClient.WriteTransaction(r.Context(), transaction); err != nil {
				return err
			}

			createLpa := lpastore.CreateLpa{
				LpaType:  lpadata.LpaTypePropertyAndAffairs,
				Channel:  lpadata.ChannelPaper,
				Language: lpaLanguage,
				Donor: lpadata.Donor{
					UID:        actoruid.New(),
					FirstNames: "Feed",
					LastName:   "Bundlaaaa",
					Address: place.Address{
						Line1:      "74 Cloob Close",
						TownOrCity: "Mahhhhhhhhhh",
						Country:    "GB",
					},
					DateOfBirth:               date.New("1970", "1", "24"),
					Email:                     "nobody@not.a.real.domain",
					ContactLanguagePreference: localize.En,
				},
				Attorneys: []lpadata.Attorney{
					{
						UID:        actoruid.New(),
						FirstNames: "Herman",
						LastName:   "Seakrest",
						Address: place.Address{
							Line1:      "81 NighOnTimeWeBuiltIt Street",
							TownOrCity: "Mahhhhhhhhhh",
							Country:    "GB",
						},
						DateOfBirth:     date.New("1982", "07", "24"),
						Status:          lpadata.AttorneyStatusActive,
						AppointmentType: lpadata.AppointmentTypeOriginal,
						Channel:         lpadata.ChannelPaper,
					},
					{
						UID:        actoruid.New(),
						FirstNames: "Herman",
						LastName:   "Seakrest",
						Address: place.Address{
							Line1:      "81 NighOnTimeWeBuiltIt Street",
							TownOrCity: "Mahhhhhhhhhh",
							Country:    "GB",
						},
						DateOfBirth:     date.New("1982", "07", "24"),
						Status:          lpadata.AttorneyStatusActive,
						AppointmentType: lpadata.AppointmentTypeOriginal,
						Channel:         lpadata.ChannelPaper,
					},
					{
						UID:        actoruid.New(),
						FirstNames: "Herman",
						LastName:   "Seakrest",
						Address: place.Address{
							Line1:      "81 NighOnTimeWeBuiltIt Street",
							TownOrCity: "Mahhhhhhhhhh",
							Country:    "GB",
						},
						DateOfBirth:     date.New("1982", "07", "24"),
						Status:          lpadata.AttorneyStatusInactive,
						AppointmentType: lpadata.AppointmentTypeReplacement,
						Channel:         lpadata.ChannelPaper,
					},
					{
						UID:        actoruid.New(),
						FirstNames: "Herman",
						LastName:   "Seakrest",
						Address: place.Address{
							Line1:      "81 NighOnTimeWeBuiltIt Street",
							TownOrCity: "Mahhhhhhhhhh",
							Country:    "GB",
						},
						DateOfBirth:     date.New("1982", "07", "24"),
						Status:          lpadata.AttorneyStatusInactive,
						AppointmentType: lpadata.AppointmentTypeReplacement,
						Channel:         lpadata.ChannelPaper,
					},
				},
				CertificateProvider: lpadata.CertificateProvider{
					UID:        actoruid.New(),
					FirstNames: "Vone",
					LastName:   "Spust",
					Address: place.Address{
						Line1:      "122111 Zonnington Way",
						TownOrCity: "Mahhhhhhhhhh",
						Country:    "GB",
					},
					Channel: channel,
					Email:   "a@example.com",
					Phone:   phone,
				},
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
			}

			if channel.IsPaper() {
				now := time.Now()
				createLpa.CertificateProvider.SignedAt = &now
			}

			if lpaType == "personal-welfare" {
				createLpa.LpaType = lpadata.LpaTypePersonalWelfare
			}

			if err := lpaStoreClient.SendLpa(r.Context(), donorDetails.LpaUID, createLpa); err != nil {
				return err
			}

		} else if isSupported {
			supporterCtx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, Email: testEmail})

			member, err := memberStore.Create(supporterCtx, random.AlphaNumeric(12), random.AlphaNumeric(12))
			if err != nil {
				return err
			}

			org, err := organisationStore.Create(supporterCtx, member, random.AlphaNumeric(12))
			if err != nil {
				return err
			}

			orgSession := &appcontext.Session{SessionID: donorSessionID, OrganisationID: org.ID}
			donorDetails, err = organisationStore.CreateLPA(appcontext.ContextWithSession(r.Context(), orgSession))
			if err != nil {
				return err
			}

			if err := donorStore.Link(appcontext.ContextWithSession(r.Context(), orgSession), accesscodedata.Link{
				LpaKey:      donorDetails.PK,
				LpaOwnerKey: donorDetails.SK,
				LpaUID:      donorDetails.LpaUID,
			}, donorDetails.Donor.Email); err != nil {
				return err
			}
		} else {
			donorDetails, err = donorStore.Create(appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID}))
			if err != nil {
				return err
			}
		}

		var (
			donorCtx               = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: donorSessionID, LpaID: donorDetails.LpaID})
			certificateProviderCtx = appcontext.ContextWithSession(r.Context(), &appcontext.Session{SessionID: certificateProviderSessionID, LpaID: donorDetails.LpaID})
		)

		if !isPaperDonor {
			donorDetails.Donor = makeDonor(donorEmail, "Sam", "Smith")
			donorDetails.Donor.LpaLanguagePreference = lpaLanguage

			donorDetails.Type = lpadata.LpaTypePropertyAndAffairs
			if lpaType == "personal-welfare" {
				donorDetails.Type = lpadata.LpaTypePersonalWelfare
				donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenCapacityLost
				donorDetails.LifeSustainingTreatmentOption = lpadata.LifeSustainingTreatmentOptionA
			} else {
				donorDetails.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenHasCapacity
			}

			donorDetails.Restrictions = makeRestriction(donorDetails)

			donorDetails.Attorneys = donordata.Attorneys{
				Attorneys: []donordata.Attorney{makeAttorney(attorneyNames[0]), makeAttorney(attorneyNames[1])},
			}

			donorDetails.AttorneyDecisions = donordata.AttorneyDecisions{How: lpadata.JointlyAndSeverally}

			donorDetails.CertificateProvider = makeCertificateProvider()
			if email != "" {
				donorDetails.CertificateProvider.Email = email
			}

			donorDetails.CertificateProvider.Mobile = phone

			if asProfessionalCertificateProvider {
				donorDetails.CertificateProvider.Relationship = lpadata.Professionally
			}

			if progress >= slices.Index(progressValues, "paid") {
				donorDetails.PaymentDetails = append(donorDetails.PaymentDetails, donordata.Payment{
					PaymentReference: random.AlphaNumeric(12),
					PaymentID:        random.AlphaNumeric(12),
				})
				donorDetails.Tasks.PayForLpa = task.PaymentStateCompleted
			}

			if progress >= slices.Index(progressValues, "signedByDonor") {
				donorDetails.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
				donorDetails.Tasks.SignTheLpa = task.StateCompleted
				donorDetails.SignedAt = testNow
				donorDetails.WitnessedByCertificateProviderAt = testNow
			}

			if err := donorStore.Put(donorCtx, donorDetails); err != nil {
				return err
			}

			if useRealUID {
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
			} else {
				donorDetails.LpaUID = makeUID()
			}

			if !donorDetails.SignedAt.IsZero() {
				if err := lpaStoreClient.SendLpa(donorCtx, donorDetails.LpaUID, lpastore.CreateLpaFromDonorProvided(donorDetails)); err != nil {
					return err
				}
			}
		}

		// should only be used in tests as otherwise people can read their emails...
		if accessCode != "" {
			accessCodeSender.UseTestCode(accessCode)
		}

		if email != "" {
			accessCodeSender.SendCertificateProviderInvite(donorCtx, appcontext.Data{
				SessionID: donorSessionID,
				LpaID:     donorDetails.LpaID,
				Localizer: appData.Localizer,
			}, donorDetails)

			switch redirect {
			case "":
				redirect = page.PathDashboard.Format()
			case page.PathCertificateProviderStart.Format():
				redirect = page.PathCertificateProviderStart.Format()
			case page.PathCertificateProviderEnterAccessCodeOptOut.Format():
				redirect = page.PathCertificateProviderEnterAccessCodeOptOut.Format()
			default:
				redirect = "/certificate-provider/" + donorDetails.LpaID + redirect
			}

			http.Redirect(w, r, redirect, http.StatusFound)
			return nil
		}

		certificateProvider, err := createCertificateProvider(certificateProviderCtx, accessCodeStore, certificateProviderStore, donorDetails)
		if err != nil {
			return err
		}

		if progress >= slices.Index(progressValues, "confirmYourDetails") {
			certificateProvider.DateOfBirth = date.New("1990", "1", "2")
			certificateProvider.ContactLanguagePreference = localize.En
			certificateProvider.Tasks.ConfirmYourDetails = task.StateCompleted

			if asProfessionalCertificateProvider {
				certificateProvider.HomeAddress = place.Address{
					Line1:      "6 RICHMOND PLACE",
					Line2:      "KINGS HEATH",
					Line3:      "WEST MIDLANDS",
					TownOrCity: "BIRMINGHAM",
					Postcode:   "B14 7ED",
				}
			}
		}

		if progress >= slices.Index(progressValues, "confirmYourIdentity") {
			var userData identity.UserData

			switch idStatus {
			case "mismatch":
				userData = identity.UserData{
					Status:      identity.StatusConfirmed,
					CheckedAt:   time.Now(),
					FirstNames:  "a",
					LastName:    "b",
					DateOfBirth: certificateProvider.DateOfBirth,
				}
				certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStatePending
			case "post-office":
				certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStatePending
			default:
				userData = identity.UserData{
					Status:      identity.StatusConfirmed,
					CheckedAt:   time.Now(),
					FirstNames:  donorDetails.CertificateProvider.FirstNames,
					LastName:    donorDetails.CertificateProvider.LastName,
					DateOfBirth: certificateProvider.DateOfBirth,
				}
				certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted
			}

			certificateProvider.IdentityUserData = userData
		}

		if progress >= slices.Index(progressValues, "provideYourCertificate") {
			certificateProvider.SignedAt = time.Now()
			certificateProvider.Tasks.ProvideTheCertificate = task.StateCompleted
		}

		if err := certificateProviderStore.Put(certificateProviderCtx, certificateProvider); err != nil {
			return err
		}

		switch redirect {
		case "":
			redirect = page.PathDashboard.Format()
		case page.PathCertificateProviderStart.Format():
			redirect = page.PathCertificateProviderStart.Format()
		case page.PathCertificateProviderEnterAccessCodeOptOut.Format():
			redirect = page.PathCertificateProviderEnterAccessCodeOptOut.Format()
		default:
			redirect = "/certificate-provider/" + donorDetails.LpaID + redirect
		}

		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
