package task

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestProgressToSlice(t *testing.T) {
	progress := Progress{
		Paid:                      ProgressTask{Label: "Paid translation"},
		ConfirmedID:               ProgressTask{Label: "ConfirmedID translation"},
		DonorSigned:               ProgressTask{Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{Label: "AttorneysSigned translation"},
		StatutoryWaitingPeriod:    ProgressTask{Label: "StatutoryWaitingPeriod translation"},
		Registered:                ProgressTask{Label: "LpaRegistered translation"},
	}

	assert.Equal(t, []ProgressTask{
		progress.Paid,
		progress.ConfirmedID,
		progress.DonorSigned,
		progress.CertificateProviderSigned,
		progress.AttorneysSigned,
		progress.StatutoryWaitingPeriod,
		progress.Registered,
	}, progress.ToSlice())
}

func TestProgressTrackerProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	initialProgress := Progress{
		Paid:                      ProgressTask{Label: "Paid translation"},
		ConfirmedID:               ProgressTask{Label: "ConfirmedID translation"},
		DonorSigned:               ProgressTask{Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{Label: "AttorneysSigned translation"},
		StatutoryWaitingPeriod:    ProgressTask{Label: "StatutoryWaitingPeriod translation"},
		Registered:                ProgressTask{Label: "LpaRegistered translation"},
	}

	testCases := map[string]struct {
		lpa              *lpadata.Lpa
		expectedProgress func() Progress
	}{
		"initial state": {
			lpa: &lpadata.Lpa{},
			expectedProgress: func() Progress {
				return initialProgress
			},
		},
		"paid": {
			lpa: &lpadata.Lpa{
				Paid: true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true

				return progress
			},
		},
		"lpa signed": {
			lpa: &lpadata.Lpa{
				Donor:                            lpadata.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.Done = true

				return progress
			},
		},
		"certificate provider signed": {
			lpa: &lpadata.Lpa{
				Paid:  true,
				Donor: lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true

				return progress
			},
		},
		"attorneys signed": {
			lpa: &lpadata.Lpa{
				Paid:  true,
				Donor: lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}, {UID: uid2, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true

				return progress
			},
		},
		"submitted": {
			lpa: &lpadata.Lpa{
				Paid:  true,
				Donor: lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Submitted:                        true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true

				return progress
			},
		},
		"statutory waiting period": {
			lpa: &lpadata.Lpa{
				Paid:  true,
				Donor: lpadata.Donor{FirstNames: "a", LastName: "b"},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt}}},
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Submitted:                        true,
				StatutoryWaitingPeriodAt:         lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true
				progress.StatutoryWaitingPeriod.Done = true

				return progress
			},
		},
		"registered": {
			lpa: &lpadata.Lpa{
				Paid:                             true,
				Donor:                            lpadata.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid1, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Submitted:                true,
				StatutoryWaitingPeriodAt: lpaSignedAt,
				RegisteredAt:             lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true
				progress.StatutoryWaitingPeriod.Done = true
				progress.Registered.Done = true

				return progress
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("lpaPaidFor").
				Return("Paid translation").
				Once()
			localizer.EXPECT().
				T("yourIdentityConfirmed").
				Return("ConfirmedID translation").
				Once()
			localizer.EXPECT().
				T("lpaSignedByYou").
				Return("DonorSigned translation").
				Once()
			localizer.EXPECT().
				T("lpaCertificateProvided").
				Return("CertificateProviderSigned translation").
				Once()
			localizer.EXPECT().
				T("lpaSignedByAllAttorneys").
				Return("AttorneysSigned translation").
				Once()
			localizer.EXPECT().
				T("opgStatutoryWaitingPeriodBegins").
				Return("StatutoryWaitingPeriod translation").
				Once()
			localizer.EXPECT().
				T("lpaRegisteredByOpg").
				Return("LpaRegistered translation").
				Once()

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
		Paid:                      ProgressTask{Label: "Paid translation"},
		ConfirmedID:               ProgressTask{Label: "ConfirmedID translation"},
		DonorSigned:               ProgressTask{Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{Label: "AttorneysSigned translation"},
		StatutoryWaitingPeriod:    ProgressTask{Label: "StatutoryWaitingPeriod translation"},
		Registered:                ProgressTask{Label: "LpaRegistered translation"},
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
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					T("lpaCertificateProvided").
					Return("CertificateProviderSigned translation")
			},
		},
		"paid": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor:               lpadata.Donor{FirstNames: "a", LastName: "b"},
				Attorneys:           lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
				},
				Paid: true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
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
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
				},
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				Paid:      true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
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
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
				},
				Attorneys:                        lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				Paid:                             true,
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true
				progress.DonorSigned.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
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
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Paid:                             true,
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
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
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Paid:                             true,
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
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
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Paid:                             true,
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Submitted:                        true,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
			},
		},
		"statutory waiting period": {
			lpa: &lpadata.Lpa{
				IsOrganisationDonor: true,
				Donor: lpadata.Donor{
					FirstNames:    "a",
					LastName:      "b",
					DateOfBirth:   dateOfBirth,
					IdentityCheck: lpadata.IdentityCheck{CheckedAt: time.Now()},
				},
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Paid:                             true,
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Submitted:                        true,
				StatutoryWaitingPeriodAt:         lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true
				progress.StatutoryWaitingPeriod.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
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
				Attorneys: lpadata.Attorneys{Attorneys: []lpadata.Attorney{{UID: uid, SignedAt: lpaSignedAt.Add(time.Minute)}}},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					SignedAt:   lpaSignedAt,
				},
				Paid:                             true,
				SignedAt:                         lpaSignedAt,
				WitnessedByCertificateProviderAt: lpaSignedAt,
				Submitted:                        true,
				StatutoryWaitingPeriodAt:         lpaSignedAt,
				RegisteredAt:                     lpaSignedAt,
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.Done = true
				progress.ConfirmedID.Done = true
				progress.DonorSigned.Done = true
				progress.CertificateProviderSigned.Done = true
				progress.AttorneysSigned.Done = true
				progress.StatutoryWaitingPeriod.Done = true
				progress.Registered.Done = true

				return progress
			},
			setupLocalizer: func(localizer *mockLocalizer) {
				localizer.EXPECT().
					Format("lpaCertificateProvidedBy",
						map[string]any{"CertificateProviderFullName": "c d"}).
					Return("CertificateProviderSigned translation")
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("lpaPaidFor").
				Return("Paid translation")
			localizer.EXPECT().
				Format("donorsIdentityConfirmed",
					map[string]any{"DonorFullName": "a b"}).
				Return("ConfirmedID translation")
			localizer.EXPECT().
				Format("lpaSignedByDonor",
					map[string]any{"DonorFullName": "a b"}).
				Return("DonorSigned translation")
			localizer.EXPECT().
				T("lpaSignedByAllAttorneys").
				Return("AttorneysSigned translation")
			localizer.EXPECT().
				T("opgStatutoryWaitingPeriodBegins").
				Return("StatutoryWaitingPeriod translation")
			localizer.EXPECT().
				Format("donorsLpaRegisteredByOpg",
					map[string]any{"DonorFullName": "a b"}).
				Return("LpaRegistered translation")

			if tc.setupLocalizer != nil {
				tc.setupLocalizer(localizer)
			}

			progressTracker := ProgressTracker{Localizer: localizer}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.lpa))
		})
	}
}
