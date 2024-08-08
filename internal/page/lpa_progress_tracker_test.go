package page

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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
			p.NoticesOfIntentSent.State = task.StateCompleted

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
			p.NoticesOfIntentSent.State = task.StateCompleted

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
				Paid:                      ProgressTask{State: task.StateNotStarted, Label: "Paid translation"},
				ConfirmedID:               ProgressTask{State: task.StateNotStarted, Label: "ConfirmedID translation"},
				DonorSigned:               ProgressTask{State: task.StateInProgress, Label: "DonorSigned translation"},
				CertificateProviderSigned: ProgressTask{State: task.StateNotStarted, Label: "CertificateProviderSigned translation"},
				AttorneysSigned:           ProgressTask{State: task.StateNotStarted, Label: "AttorneysSigned translation"},
				LpaSubmitted:              ProgressTask{State: task.StateNotStarted, Label: "LpaSubmitted translation"},
				NoticesOfIntentSent:       ProgressTask{State: task.StateNotStarted, Label: "NoticesOfIntentSent translation"},
				StatutoryWaitingPeriod:    ProgressTask{State: task.StateNotStarted, Label: "StatutoryWaitingPeriod translation"},
				LpaRegistered:             ProgressTask{State: task.StateNotStarted, Label: "LpaRegistered translation"},
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
		Paid:                      ProgressTask{State: task.StateNotStarted, Label: ""},
		ConfirmedID:               ProgressTask{State: task.StateNotStarted, Label: ""},
		DonorSigned:               ProgressTask{State: task.StateInProgress, Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{State: task.StateNotStarted, Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{State: task.StateNotStarted, Label: "AttorneysSigned translation"},
		LpaSubmitted:              ProgressTask{State: task.StateNotStarted, Label: "LpaSubmitted translation"},
		StatutoryWaitingPeriod:    ProgressTask{State: task.StateNotStarted, Label: "StatutoryWaitingPeriod translation"},
		LpaRegistered:             ProgressTask{State: task.StateNotStarted, Label: "LpaRegistered translation"},
	}

	testCases := map[string]struct {
		lpa              *lpadata.Lpa
		expectedProgress func() Progress
		setupLocalizer   func(*mockLocalizer)
	}{
		"initial state": {
			lpa: &lpadata.Lpa{
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
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
			lpa: &lpadata.Lpa{
				CertificateProvider: lpadata.CertificateProvider{FirstNames: "A", LastName: "B"},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
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
			lpa: &lpadata.Lpa{
				Donor:     lpadata.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateInProgress

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
			lpa: &lpadata.Lpa{
				Paid:                true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				SignedAt:            lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateInProgress

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
			lpa: &lpadata.Lpa{
				Paid:                true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}, {UID: uid2, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				SignedAt:            lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateInProgress

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
			lpa: &lpadata.Lpa{
				Paid:                true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
				SignedAt:            lpaSignedAt,
				Submitted:           true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateCompleted

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
			lpa: &lpadata.Lpa{
				Paid:                true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
				SignedAt:            lpaSignedAt,
				Submitted:           true,
				PerfectAt:           lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateCompleted
				progress.NoticesOfIntentSent.State = task.StateCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.StatutoryWaitingPeriod.State = task.StateInProgress

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
			lpa: &lpadata.Lpa{
				Paid:                true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:            lpaSignedAt,
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Submitted:           true,
				PerfectAt:           lpaSignedAt,
				RegisteredAt:        lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateCompleted
				progress.NoticesOfIntentSent.State = task.StateCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.StatutoryWaitingPeriod.State = task.StateCompleted
				progress.LpaRegistered.State = task.StateCompleted

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
		Paid:                      ProgressTask{State: task.StateInProgress, Label: "Paid translation"},
		ConfirmedID:               ProgressTask{State: task.StateNotStarted, Label: "ConfirmedID translation"},
		DonorSigned:               ProgressTask{State: task.StateNotStarted, Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{State: task.StateNotStarted, Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{State: task.StateNotStarted, Label: "AttorneysSigned translation"},
		LpaSubmitted:              ProgressTask{State: task.StateNotStarted, Label: "LpaSubmitted translation"},
		NoticesOfIntentSent:       ProgressTask{State: task.StateNotStarted},
		StatutoryWaitingPeriod:    ProgressTask{State: task.StateNotStarted, Label: "StatutoryWaitingPeriod translation"},
		LpaRegistered:             ProgressTask{State: task.StateNotStarted, Label: "LpaRegistered translation"},
	}

	testCases := map[string]struct {
		lpa              *lpadata.Lpa
		expectedProgress func() Progress
		setupLocalizer   func(*mockLocalizer)
	}{
		"initial state": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
			},
			expectedProgress: func() Progress {
				return initialProgress
			},
		},
		"paid": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				Paid:                true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateInProgress

				return progress
			},
		},
		"confirmed ID": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				Paid:      true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateInProgress

				return progress
			},
		},
		"donor signed": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				Paid:      true,
				SignedAt:  lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateInProgress

				return progress
			},
		},
		"certificate provider signed": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                true,
				SignedAt:            lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateInProgress

				return progress
			},
		},
		"attorneys signed": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                true,
				SignedAt:            lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateInProgress

				return progress
			},
		},
		"submitted": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                true,
				SignedAt:            lpaSignedAt,
				Submitted:           true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateCompleted

				return progress
			},
		},
		"perfect": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                true,
				SignedAt:            lpaSignedAt,
				Submitted:           true,
				PerfectAt:           lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.NoticesOfIntentSent.State = task.StateCompleted
				progress.StatutoryWaitingPeriod.State = task.StateInProgress

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
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{SignedAt: lpaSignedAt},
				Paid:                true,
				SignedAt:            lpaSignedAt,
				Submitted:           true,
				PerfectAt:           lpaSignedAt,
				RegisteredAt:        lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = task.StateCompleted
				progress.ConfirmedID.State = task.StateCompleted
				progress.DonorSigned.State = task.StateCompleted
				progress.CertificateProviderSigned.State = task.StateCompleted
				progress.AttorneysSigned.State = task.StateCompleted
				progress.LpaSubmitted.State = task.StateCompleted
				progress.NoticesOfIntentSent.Label = "NoticesOfIntentSent translation"
				progress.NoticesOfIntentSent.State = task.StateCompleted
				progress.StatutoryWaitingPeriod.State = task.StateCompleted
				progress.LpaRegistered.State = task.StateCompleted

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
