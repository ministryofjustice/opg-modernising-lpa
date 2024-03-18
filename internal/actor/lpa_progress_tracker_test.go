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
		DonorSigned:               ProgressTask{State: TaskInProgress, Label: "You’ve signed your LPA"},
		CertificateProviderSigned: ProgressTask{State: TaskNotStarted, Label: "Your certificate provider has provided their certificate"},
		AttorneysSigned:           ProgressTask{State: TaskNotStarted, Label: "Your attorney has signed your LPA"},
		LpaSubmitted:              ProgressTask{State: TaskNotStarted, Label: "We have received your LPA"},
		StatutoryWaitingPeriod:    ProgressTask{State: TaskNotStarted, Label: "Your 4-week waiting period has started"},
		LpaRegistered:             ProgressTask{State: TaskNotStarted, Label: "Your LPA has been registered"},
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
				Attorneys: Attorneys{Attorneys: []Attorney{{}}},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				return initialProgress
			},
		},
		"initial state - with certificate provider name": {
			donor: &DonorProvidedDetails{
				CertificateProvider: CertificateProvider{FirstNames: "A", LastName: "B"},
				Attorneys:           Attorneys{Attorneys: []Attorney{{}}},
			},
			certificateProvider: &CertificateProviderProvidedDetails{},
			expectedProgress: func() Progress {
				progress := initialProgress
				progress.CertificateProviderSigned.Label = "A B has provided their certificate"

				return progress
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
				progress.AttorneysSigned.Label = "Your attorneys have signed your LPA"
				progress.LpaSubmitted.State = TaskInProgress

				return progress
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
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
