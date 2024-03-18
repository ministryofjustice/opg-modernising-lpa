package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
)

func TestProgressTrackerProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	initialProgress := Progress{
		Paid:                      ProgressTask{State: TaskNotStarted, Label: ""},
		ConfirmedID:               ProgressTask{State: TaskNotStarted, Label: ""},
		DonorSigned:               ProgressTask{State: TaskInProgress, Label: "DonorSigned translation"},
		CertificateProviderSigned: ProgressTask{State: TaskNotStarted, Label: "CertificateProviderSigned translation"},
		AttorneysSigned:           ProgressTask{State: TaskNotStarted, Label: "AttorneysSigned translation"},
		LpaSubmitted:              ProgressTask{State: TaskNotStarted, Label: "LpaSubmitted translation"},
		StatutoryWaitingPeriod:    ProgressTask{State: TaskNotStarted, Label: "StatutoryWaitingPeriod translation"},
		LpaRegistered:             ProgressTask{State: TaskNotStarted, Label: "LpaRegistered translation"},
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
		donor               *DonorProvidedDetails
		certificateProvider *CertificateProviderProvidedDetails
		attorneys           []*AttorneyProvidedDetails
		expectedProgress    func() Progress
		expectedLocalizer   func() *mockLocalizer
	}{
		"initial state": {
			donor: &DonorProvidedDetails{
				Attorneys: Attorneys{Attorneys: []Attorney{{}}},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				return initialProgress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"initial state - with certificate provider name": {
			donor: &DonorProvidedDetails{
				CertificateProvider: CertificateProvider{FirstNames: "A", LastName: "B"},
				Attorneys:           Attorneys{Attorneys: []Attorney{{}}},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
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
			donor: &DonorProvidedDetails{
				Donor:     Donor{FirstNames: "a", LastName: "b"},
				Attorneys: Attorneys{Attorneys: []Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"certificate provider signed": {
			donor: &DonorProvidedDetails{
				Tasks:     DonorTasks{PayForLpa: PaymentTaskCompleted},
				Donor:     Donor{FirstNames: "a", LastName: "b"},
				Attorneys: Attorneys{Attorneys: []Attorney{{}}},
				SignedAt:  lpaSignedAt,
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"attorneys signed": {
			donor: &DonorProvidedDetails{
				Tasks:     DonorTasks{PayForLpa: PaymentTaskCompleted},
				Donor:     Donor{FirstNames: "a", LastName: "b"},
				SignedAt:  lpaSignedAt,
				Attorneys: Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskCompleted
				progress.LpaSubmitted.State = TaskInProgress

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
			donor: &DonorProvidedDetails{
				Tasks:       DonorTasks{PayForLpa: PaymentTaskCompleted},
				Donor:       Donor{FirstNames: "a", LastName: "b"},
				SignedAt:    lpaSignedAt,
				Attorneys:   Attorneys{Attorneys: []Attorney{{UID: uid1}}},
				SubmittedAt: lpaSignedAt.Add(time.Hour),
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskCompleted
				progress.LpaSubmitted.State = TaskCompleted
				progress.StatutoryWaitingPeriod.State = TaskInProgress

				return progress
			},
			expectedLocalizer: func() *mockLocalizer { return localizerFn() },
		},
		"registered": {
			donor: &DonorProvidedDetails{
				Tasks:        DonorTasks{PayForLpa: PaymentTaskCompleted},
				Donor:        Donor{FirstNames: "a", LastName: "b"},
				SignedAt:     lpaSignedAt,
				Attorneys:    Attorneys{Attorneys: []Attorney{{UID: uid1}}},
				SubmittedAt:  lpaSignedAt.Add(time.Hour),
				RegisteredAt: lpaSignedAt.Add(2 * time.Hour),
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskCompleted
				progress.LpaSubmitted.State = TaskCompleted
				progress.StatutoryWaitingPeriod.State = TaskCompleted
				progress.LpaRegistered.State = TaskCompleted

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
		Paid:                      ProgressTask{State: TaskInProgress, Label: "a b has paid"},
		ConfirmedID:               ProgressTask{State: TaskNotStarted, Label: "a b has confirmed their identity"},
		DonorSigned:               ProgressTask{State: TaskNotStarted, Label: "a b has signed the LPA"},
		CertificateProviderSigned: ProgressTask{State: TaskNotStarted, Label: "The certificate provider has provided their certificate"},
		AttorneysSigned:           ProgressTask{State: TaskNotStarted, Label: "All attorneys have signed the LPA"},
		LpaSubmitted:              ProgressTask{State: TaskNotStarted, Label: "OPG has received the LPA"},
		StatutoryWaitingPeriod:    ProgressTask{State: TaskNotStarted, Label: "The 4-week waiting period has started"},
		LpaRegistered:             ProgressTask{State: TaskNotStarted, Label: "The LPA has been registered"},
	}
	bundle, err := localize.NewBundle("../../lang/en.json")
	if err != nil {
		t.Error("error creating bundle")
	}

	localizer := bundle.For(localize.En)
	progressTracker := ProgressTracker{Localizer: localizer}

	testCases := map[string]struct {
		donor               *DonorProvidedDetails
		certificateProvider *CertificateProviderProvidedDetails
		attorneys           []*AttorneyProvidedDetails
		expectedProgress    func() Progress
	}{
		"initial state": {
			donor: &DonorProvidedDetails{
				SK:        "ORGANISATION#123",
				Donor:     Donor{FirstNames: "a", LastName: "b"},
				Attorneys: Attorneys{Attorneys: []Attorney{{}}},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				return initialProgress
			},
		},
		"paid": {
			donor: &DonorProvidedDetails{
				SK:        "ORGANISATION#123",
				Donor:     Donor{FirstNames: "a", LastName: "b"},
				Attorneys: Attorneys{Attorneys: []Attorney{{}}},
				Tasks:     DonorTasks{PayForLpa: PaymentTaskCompleted},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskInProgress

				return progress
			},
		},
		"confirmed ID": {
			donor: &DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             Attorneys{Attorneys: []Attorney{{}}},
				Tasks:                 DonorTasks{PayForLpa: PaymentTaskCompleted},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskCompleted
				progress.DonorSigned.State = TaskInProgress

				return progress
			},
		},
		"donor signed": {
			donor: &DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             Attorneys{Attorneys: []Attorney{{}}},
				Tasks:                 DonorTasks{PayForLpa: PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskCompleted
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskInProgress

				return progress
			},
		},
		"certificate provider signed": {
			donor: &DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             Attorneys{Attorneys: []Attorney{{}}},
				Tasks:                 DonorTasks{PayForLpa: PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskCompleted
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskInProgress

				return progress
			},
		},
		"attorneys signed": {
			donor: &DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             Attorneys{Attorneys: []Attorney{{UID: uid}}},
				Tasks:                 DonorTasks{PayForLpa: PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskCompleted
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskCompleted
				progress.LpaSubmitted.State = TaskInProgress

				return progress
			},
		},
		"submitted": {
			donor: &DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             Attorneys{Attorneys: []Attorney{{UID: uid}}},
				Tasks:                 DonorTasks{PayForLpa: PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
				SubmittedAt:           lpaSignedAt.Add(time.Hour),
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskCompleted
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskCompleted
				progress.LpaSubmitted.State = TaskCompleted
				progress.StatutoryWaitingPeriod.State = TaskInProgress

				return progress
			},
		},
		"registered": {
			donor: &DonorProvidedDetails{
				SK:                    "ORGANISATION#123",
				Donor:                 Donor{FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
				Attorneys:             Attorneys{Attorneys: []Attorney{{UID: uid}}},
				Tasks:                 DonorTasks{PayForLpa: PaymentTaskCompleted},
				SignedAt:              lpaSignedAt,
				SubmittedAt:           lpaSignedAt.Add(time.Hour),
				RegisteredAt:          lpaSignedAt.Add(2 * time.Hour),
			},
			certificateProvider: &CertificateProviderProvidedDetails{Certificate: Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.Paid.State = TaskCompleted
				progress.ConfirmedID.State = TaskCompleted
				progress.DonorSigned.State = TaskCompleted
				progress.CertificateProviderSigned.State = TaskCompleted
				progress.AttorneysSigned.State = TaskCompleted
				progress.LpaSubmitted.State = TaskCompleted
				progress.StatutoryWaitingPeriod.State = TaskCompleted
				progress.LpaRegistered.State = TaskCompleted

				return progress
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedProgress(), progressTracker.Progress(tc.donor, tc.certificateProvider, tc.attorneys))
		})
	}
}
