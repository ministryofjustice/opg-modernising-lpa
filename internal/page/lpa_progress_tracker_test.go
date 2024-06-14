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

func TestProgressToSlice(t *testing.T) {
	testcases := map[string]func(Progress) (Progress, []ProgressTask){
		"donor": func(p Progress) (Progress, []ProgressTask) {
			return p, []ProgressTask{
				p.DonorSigned,
				p.CertificateProviderSigned,
				p.AttorneysSigned,
				p.LpaSubmitted,
				p.StatutoryWaitingPeriod,
				p.LpaRegistered,
			}
		},
		"organisation": func(p Progress) (Progress, []ProgressTask) {
			p.isOrganisation = true

			return p, []ProgressTask{
				p.Paid,
				p.ConfirmedID,
				p.DonorSigned,
				p.CertificateProviderSigned,
				p.AttorneysSigned,
				p.LpaSubmitted,
				p.StatutoryWaitingPeriod,
				p.LpaRegistered,
			}
		},
		"donor notices sent": func(p Progress) (Progress, []ProgressTask) {
			p.NoticesOfIntentSent.State = actor.TaskCompleted

			return p, []ProgressTask{
				p.DonorSigned,
				p.CertificateProviderSigned,
				p.AttorneysSigned,
				p.LpaSubmitted,
				p.NoticesOfIntentSent,
				p.StatutoryWaitingPeriod,
				p.LpaRegistered,
			}
		},
		"organisation notices sent": func(p Progress) (Progress, []ProgressTask) {
			p.isOrganisation = true
			p.NoticesOfIntentSent.State = actor.TaskCompleted

			return p, []ProgressTask{
				p.Paid,
				p.ConfirmedID,
				p.DonorSigned,
				p.CertificateProviderSigned,
				p.AttorneysSigned,
				p.LpaSubmitted,
				p.NoticesOfIntentSent,
				p.StatutoryWaitingPeriod,
				p.LpaRegistered,
			}
		},
	}

	for name, fn := range testcases {
		t.Run(name, func(t *testing.T) {
			progress, slice := fn(Progress{
				Paid:                      ProgressTask{State: actor.TaskNotStarted, Label: "Paid translation"},
				ConfirmedID:               ProgressTask{State: actor.TaskNotStarted, Label: "ConfirmedID translation"},
				DonorSigned:               ProgressTask{State: actor.TaskInProgress, Label: "DonorSigned translation"},
				CertificateProviderSigned: ProgressTask{State: actor.TaskNotStarted, Label: "CertificateProviderSigned translation"},
				AttorneysSigned:           ProgressTask{State: actor.TaskNotStarted, Label: "AttorneysSigned translation"},
				LpaSubmitted:              ProgressTask{State: actor.TaskNotStarted, Label: "LpaSubmitted translation"},
				NoticesOfIntentSent:       ProgressTask{State: actor.TaskNotStarted, Label: "NoticesOfIntentSent translation"},
				StatutoryWaitingPeriod:    ProgressTask{State: actor.TaskNotStarted, Label: "StatutoryWaitingPeriod translation"},
				LpaRegistered:             ProgressTask{State: actor.TaskNotStarted, Label: "LpaRegistered translation"},
			})

			assert.Equal(t, slice, progress.ToSlice())
		})
	}
}

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

	testCases := map[string]struct {
		lpa              *lpastore.Lpa
		expectedProgress func() Progress
		setupLocalizer   func(*mockLocalizer)
	}{
		"initial state": {
			lpa: &lpastore.Lpa{
				Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
			},
		},
		"initial state with certificate provider name": {
			lpa: &lpastore.Lpa{
				CertificateProvider: lpastore.CertificateProvider{FirstNames: "A", LastName: "B"},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("certificateProviderHasDeclared", map[string]interface{}{
						"CertificateProviderFullName": "A B",
					}).
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
			},
		},
		"lpa signed": {
			lpa: &lpastore.Lpa{
				Donor:     lpastore.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskInProgress

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
			},
		},
		"certificate provider signed": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
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
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
			},
		},
		"attorneys signed": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
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
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 2).
					Return("AttorneysSigned translation")
			},
		},
		"submitted": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
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

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
			},
		},
		"perfect": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
				SignedAt:            lpaSignedAt,
				Submitted:           true,
				PerfectAt:           lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.NoticesOfIntentSent.State = actor.TaskCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.StatutoryWaitingPeriod.State = actor.TaskInProgress

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
				localizer.EXPECT().
					Format("weSentAnEmailYourLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
					Return("NoticesOfIntentSent translation")
				localizer.EXPECT().
					FormatDate(lpaSignedAt).
					Return("perfect-on")
			},
		},
		"registered": {
			lpa: &lpastore.Lpa{
				Paid:                true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:            lpaSignedAt,
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Submitted:           true,
				PerfectAt:           lpaSignedAt,
				RegisteredAt:        lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.NoticesOfIntentSent.State = actor.TaskCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.StatutoryWaitingPeriod.State = actor.TaskCompleted
				progress.LpaRegistered.State = actor.TaskCompleted

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("yourCertificateProviderHasDeclared").
					Return("CertificateProviderSigned translation")
				localizer.EXPECT().
					Count("attorneysHaveDeclared", 1).
					Return("AttorneysSigned translation")
				localizer.EXPECT().
					Format("weSentAnEmailYourLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
					Return("NoticesOfIntentSent translation")
				localizer.EXPECT().
					FormatDate(lpaSignedAt).
					Return("perfect-on")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("youveSignedYourLpa").
				Return("DonorSigned translation")
			localizer.EXPECT().
				T("weHaveReceivedYourLpa").
				Return("LpaSubmitted translation")
			localizer.EXPECT().
				T("yourWaitingPeriodHasStarted").
				Return("StatutoryWaitingPeriod translation")
			localizer.EXPECT().
				T("yourLpaHasBeenRegistered").
				Return("LpaRegistered translation")

			if tc.setupLocalizer != nil {
				tc.setupLocalizer(localizer)
			}

			progressTracker := ProgressTracker{Localizer: localizer}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa))
		})
	}
}

func TestLpaProgressAsSupporter(t *testing.T) {
	dateOfBirth := date.Today()
	lpaSignedAt := time.Now()
	uid := actoruid.New()
	initialProgress := Progress{
		isOrganisation:            true,
		Paid:                      ProgressTask{State: actor.TaskInProgress, Label: "Paid translation"},
		ConfirmedID:               ProgressTask{State: actor.TaskNotStarted, Label: "ConfirmedID translation"},
		DonorSigned:               ProgressTask{State: actor.TaskNotStarted, Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{State: actor.TaskNotStarted, Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{State: actor.TaskNotStarted, Label: "AttorneysSigned translation"},
		LpaSubmitted:              ProgressTask{State: actor.TaskNotStarted, Label: "LpaSubmitted translation"},
		NoticesOfIntentSent:       ProgressTask{State: actor.TaskNotStarted},
		StatutoryWaitingPeriod:    ProgressTask{State: actor.TaskNotStarted, Label: "StatutoryWaitingPeriod translation"},
		LpaRegistered:             ProgressTask{State: actor.TaskNotStarted, Label: "LpaRegistered translation"},
	}

	testCases := map[string]struct {
		lpa              *lpastore.Lpa
		expectedProgress func() Progress
		setupLocalizer   func(*mockLocalizer)
	}{
		"initial state": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor: true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
		},
		"paid": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor: true,
				Donor:               lpastore.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpastore.Attorneys{Attorneys: []lpastore.Attorney{{}}},
				Paid:                true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskInProgress

				return progress
			},
		},
		"confirmed ID": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
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
		},
		"donor signed": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
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
		},
		"certificate provider signed": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
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
		},
		"attorneys signed": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
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
		},
		"submitted": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
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

				return progress
			},
		},
		"perfect": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider:    lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
				Submitted:              true,
				PerfectAt:              lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskCompleted
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.NoticesOfIntentSent.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskInProgress

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("weSentAnEmailTheLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
					Return("NoticesOfIntentSent translation")
				localizer.EXPECT().
					FormatDate(lpaSignedAt).
					Return("perfect-on")
			},
		},
		"registered": {
			lpa: &lpastore.Lpa{
				IsOrganisationDonor:    true,
				Donor:                  lpastore.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityConfirmed: true,
				Attorneys:              lpastore.Attorneys{Attorneys: []lpastore.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider:    lpastore.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                   true,
				SignedAt:               lpaSignedAt,
				Submitted:              true,
				PerfectAt:              lpaSignedAt,
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
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.NoticesOfIntentSent.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskCompleted
				progress.LpaRegistered.State = actor.TaskCompleted

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("weSentAnEmailTheLpaIsReadyToRegister", map[string]any{"SentOn": "perfect-on"}).
					Return("NoticesOfIntentSent translation")
				localizer.EXPECT().
					FormatDate(lpaSignedAt).
					Return("perfect-on")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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

			if tc.setupLocalizer != nil {
				tc.setupLocalizer(localizer)
			}

			progressTracker := ProgressTracker{Localizer: localizer}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa))
		})
	}
}
