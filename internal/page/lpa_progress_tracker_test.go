package page

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/stretchr/testify/assert"
)

func TestProgressTrackerProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	initialProgress := Progress{
		Paid:                      ProgressTask{State: actor.TaskNotStarted, Label: ""},
		ConfirmedID:               ProgressTask{State: actor.TaskNotStarted, Label: ""},
		DonorSigned:               ProgressTask{State: actor.TaskInProgress, Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{State: actor.TaskNotStarted, Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{State: actor.TaskNotStarted, Label: "AttorneysSigned translation"},
		LpaSubmitted:              ProgressTask{State: actor.TaskNotStarted, Label: "LpaSubmitted translation"},
		StatutoryWaitingPeriod:    ProgressTask{State: actor.TaskNotStarted, Label: "StatutoryWaitingPeriod translation"},
		LpaRegistered:             ProgressTask{State: actor.TaskNotStarted, Label: "LpaRegistered translation"},
	}

	localizerFn := func() *mockLocalizer {
		localizer := newMockLocalizer(t)
		localizer.EXPECT().
			T("youveSignedYourLpa").
			Return("DonorSigned translation")
		localizer.EXPECT().
			T("yourCertificateProviderHasDeclared").
			Return("CertificateProviderSigned translation")
		localizer.EXPECT().
			Count("attorneysHaveDeclared", 1).
			Return("AttorneysSigned translation")
		localizer.EXPECT().
			T("weHaveReceivedYourLpa").
			Return("LpaSubmitted translation")
		localizer.EXPECT().
			T("yourWaitingPeriodHasStarted").
			Return("StatutoryWaitingPeriod translation")
		localizer.EXPECT().
			T("yourLpaHasBeenRegistered").
			Return("LpaRegistered translation")

		return localizer
	}

	testCases := map[string]struct {
		lpa               *lpastore.Lpa
		expectedProgress  func() Progress
		expectedLocalizer func() *mockLocalizer
	}{
		"initial state": {
			lpa: &lpastore.Lpa{
				Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"initial state - with certificate provider name": {
			lpa: &lpastore.Lpa{
				CertificateProvider: lpastore.CertificateProvider{FirstNames: "A", LastName: "B"},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
			expectedLocalizer: func() *mockLocalizer {
				localizer := newMockLocalizer(t)
				localizer.EXPECT().
					T("youveSignedYourLpa").
					Return("DonorSigned translation")
				localizer.EXPECT().
					Format(
						"certificateProviderHasDeclared", map[string]interface{}{"CertificateProviderFullName": "A B"},
					).
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
				localizer.EXPECT().
					T("weHaveReceivedYourLpa").
					Return("LpaSubmitted translation")
				localizer.EXPECT().
					T("yourWaitingPeriodHasStarted").
					Return("StatutoryWaitingPeriod translation")
				localizer.EXPECT().
					T("yourLpaHasBeenRegistered").
					Return("LpaRegistered translation")

				return localizer
			},
		},
		"lpa signed": {
			lpa: &lpastore.Lpa{
				Donor:     actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"certificate provider signed": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               actor.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				SignedAt:            lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"attorneys signed": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               actor.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}, {UID: uid2, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				SignedAt:            lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer {
				localizer := newMockLocalizer(t)
				localizer.EXPECT().
					T("youveSignedYourLpa").
					Return("DonorSigned translation")
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 2).
					Return("AttorneysSigned translation")
				localizer.EXPECT().
					T("weHaveReceivedYourLpa").
					Return("LpaSubmitted translation")
				localizer.EXPECT().
					T("yourWaitingPeriodHasStarted").
					Return("StatutoryWaitingPeriod translation")
				localizer.EXPECT().
					T("yourLpaHasBeenRegistered").
					Return("LpaRegistered translation")

				return localizer
			},
		},
		"submitted": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               actor.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
				SignedAt:            lpaSignedAt,
				Submitted:           true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"registered": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               actor.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:            lpaSignedAt,
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Submitted:           true,
				RegisteredAt:        lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskCompleted
				progress.LpaRegistered.State = actor.TaskCompleted
				progress.LpaRegistered.NotificationSentTranslation = "emailSentOnAbout translation"

				return progress
			},
			expectedLocalizer: func() *mockLocalizer {
				baseExpectationsLocalizer := localizerFn()
				baseExpectationsLocalizer.EXPECT().
					Format("emailSentOnAbout", map[string]any{"On": "Formatted dated", "About": "yourLPARegistration translation"}).
					Return("emailSentOnAbout translation")
				baseExpectationsLocalizer.EXPECT().
					FormatDate(lpaSignedAt).
					Return("Formatted dated")
				baseExpectationsLocalizer.EXPECT().
					T("yourLPARegistration").
					Return("yourLPARegistration translation")

				return baseExpectationsLocalizer
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			progressTracker := ProgressTracker{Localizer: tc.expectedLocalizer()}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa))
		})
	}
}

func TestLpaProgressAsSupporter(t *testing.T) {
	dateOfBirth := date.Today()
	lpaSignedAt := time.Now()
	uid := actoruid.New()
	initialProgress := Progress{
		Paid:                      ProgressTask{State: actor.TaskInProgress, Label: "Paid translation"},
		ConfirmedID:               ProgressTask{State: actor.TaskNotStarted, Label: "ConfirmedID translation"},
		DonorSigned:               ProgressTask{State: actor.TaskNotStarted, Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{State: actor.TaskNotStarted, Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{State: actor.TaskNotStarted, Label: "AttorneysSigned translation"},
		LpaSubmitted:              ProgressTask{State: actor.TaskNotStarted, Label: "LpaSubmitted translation"},
		StatutoryWaitingPeriod:    ProgressTask{State: actor.TaskNotStarted, Label: "StatutoryWaitingPeriod translation"},
		LpaRegistered:             ProgressTask{State: actor.TaskNotStarted, Label: "LpaRegistered translation"},
	}

	localizerFn := func() *mockLocalizer {
		localizer := newMockLocalizer(t)
		localizer.EXPECT().
			Format(
				"donorFullNameHasPaid",
				map[string]interface{}{"DonorFullName": "a b"},
			).
			Return("Paid translation")
		localizer.EXPECT().
			Format(
				"donorFullNameHasConfirmedTheirIdentity",
				map[string]interface{}{"DonorFullName": "a b"},
			).
			Return("ConfirmedID translation")
		localizer.EXPECT().
			Format(
				"donorFullNameHasSignedTheLPA",
				map[string]interface{}{"DonorFullName": "a b"},
			).
			Return("DonorSigned translation")
		localizer.EXPECT().
			T("theCertificateProviderHasDeclared").
			Return("CertificateProviderSigned translation")
		localizer.EXPECT().
			T("allAttorneysHaveSignedTheLpa").
			Return("AttorneysSigned translation")
		localizer.EXPECT().
			T("opgHasReceivedTheLPA").
			Return("LpaSubmitted translation")
		localizer.EXPECT().
			T("theWaitingPeriodHasStarted").
			Return("StatutoryWaitingPeriod translation")
		localizer.EXPECT().
			T("theLpaHasBeenRegistered").
			Return("LpaRegistered translation")

		return localizer
	}

	testCases := map[string]struct {
		lpa               *lpastore.Lpa
		expectedProgress  func() Progress
		expectedLocalizer func() *mockLocalizer
	}{
		"initial state": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor: true,
				Donor:               actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"paid": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor: true,
				Donor:               actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				Paid:                true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"confirmed ID": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				Paid:                   true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"donor signed": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"certificate provider signed": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				CertificateProvider:    lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"attorneys signed": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider:    lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"submitted": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider:    lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
				Submitted:              true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"registered": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider:    lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
				Submitted:              true,
				RegisteredAt:           lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskCompleted
				progress.LpaRegistered.State = actor.TaskCompleted
				progress.LpaRegistered.NotificationSentTranslation = "emailSentOnAbout translation"

				return progress
			},
			expectedLocalizer: func() *mockLocalizer {
				baseExpectationsLocalizer := localizerFn()
				baseExpectationsLocalizer.EXPECT().
					Format("emailSentOnAbout", map[string]any{"On": "Formatted dated", "About": "theLPARegistration translation"}).
					Return("emailSentOnAbout translation")
				baseExpectationsLocalizer.EXPECT().
					FormatDate(lpaSignedAt).
					Return("Formatted dated")
				baseExpectationsLocalizer.EXPECT().
					T("theLPARegistration").
					Return("theLPARegistration translation")

				return baseExpectationsLocalizer
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			progressTracker := ProgressTracker{Localizer: tc.expectedLocalizer()}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa))
		})
	}
}
