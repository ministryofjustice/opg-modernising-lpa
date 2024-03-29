package page

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
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
		donor               *actor.DonorProvidedDetails
		certificateProvider *actor.CertificateProviderProvidedDetails
		attorneys           []*actor.AttorneyProvidedDetails
		expectedProgress    func() Progress
		expectedLocalizer   func() *mockLocalizer
	}{
		"initial state": {
			donor: &actor.DonorProvidedDetails{
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				return initialProgress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"initial state - with certificate provider name": {
			donor: &actor.DonorProvidedDetails{
				CertificateProvider: actor.CertificateProvider{FirstNames: "A", LastName: "B"},
				Attorneys:           actor.Attorneys{Attorneys: []actor.Attorney{{}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
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
			donor: &actor.DonorProvidedDetails{
				Donor:     actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"certificate provider signed": {
			donor: &actor.DonorProvidedDetails{
				Tasks:     actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				Donor:     actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
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
			donor: &actor.DonorProvidedDetails{
				Tasks:     actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				Donor:     actor.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
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
			donor: &actor.DonorProvidedDetails{
				Tasks:       actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				Donor:       actor.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:    lpaSignedAt,
				Attorneys:   actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}}},
				SubmittedAt: lpaSignedAt.Add(time.Hour),
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
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
			donor: &actor.DonorProvidedDetails{
				Tasks:        actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				Donor:        actor.Donor{FirstNames: "a", LastName: "b"},
				SignedAt:     lpaSignedAt,
				Attorneys:    actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}}},
				SubmittedAt:  lpaSignedAt.Add(time.Hour),
				RegisteredAt: lpaSignedAt.Add(2 * time.Hour),
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = actor.TaskCompleted
				progress.CertificateProviderSigned.State = actor.TaskCompleted
				progress.AttorneysSigned.State = actor.TaskCompleted
				progress.LpaSubmitted.State = actor.TaskCompleted
				progress.StatutoryWaitingPeriod.State = actor.TaskCompleted
				progress.LpaRegistered.State = actor.TaskCompleted

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			progressTracker := ProgressTracker{Localizer: tc.expectedLocalizer()}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.donor, tc.certificateProvider, tc.attorneys))
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
		donor               *actor.DonorProvidedDetails
		certificateProvider *actor.CertificateProviderProvidedDetails
		attorneys           []*actor.AttorneyProvidedDetails
		expectedProgress    func() Progress
		expectedLocalizer   func() *mockLocalizer
	}{
		"initial state": {
			donor: &actor.DonorProvidedDetails{
				SK:        "ORGANISATION#123",
				Donor:     actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				return initialProgress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"paid": {
			donor: &actor.DonorProvidedDetails{
				SK:        "ORGANISATION#123",
				Donor:     actor.Donor{FirstNames: "a", LastName: "b"},
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				Tasks:     actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = actor.TaskCompleted
				progress.ConfirmedID.State = actor.TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"confirmed ID": {
			donor: &actor.DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				Tasks:                 actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
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
			donor: &actor.DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				Tasks:                 actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
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
			donor: &actor.DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             actor.Attorneys{Attorneys: []actor.Attorney{{}}},
				Tasks:                 actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
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
			donor: &actor.DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}},
				Tasks:                 actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
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
			donor: &actor.DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}},
				Tasks:                 actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
				SubmittedAt:           lpaSignedAt.Add(time.Hour),
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
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
			donor: &actor.DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid}}},
				Tasks:                 actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
				SubmittedAt:           lpaSignedAt.Add(time.Hour),
				RegisteredAt:          lpaSignedAt.Add(2 * time.Hour),
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
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

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			progressTracker := ProgressTracker{Localizer: tc.expectedLocalizer()}

			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.donor, tc.certificateProvider, tc.attorneys))
		})
	}
}
