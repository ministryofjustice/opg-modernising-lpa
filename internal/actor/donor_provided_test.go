package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

var address = place.Address{
	Line1:      "a",
	Line2:      "b",
	Line3:      "c",
	TownOrCity: "d",
	Postcode:   "e",
}

func TestGenerateHash(t *testing.T) {
	donor := &DonorProvidedDetails{Attorneys: Attorneys{
		Attorneys: []Attorney{
			{DateOfBirth: date.New("2000", "1", "2")},
		},
	}}
	hash, err := donor.GenerateHash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x1c746864bd23d82e), hash)

	donor.Attorneys.Attorneys[0].DateOfBirth = date.New("2001", "1", "2")
	hash, err = donor.GenerateHash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x80d2d4728e9a797a), hash)
}

func TestIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *DonorProvidedDetails
		expected bool
	}{
		"set": {
			lpa: &DonorProvidedDetails{
				Donor:                 Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b"},
			},
			expected: true,
		},
		"not ok": {
			lpa:      &DonorProvidedDetails{DonorIdentityUserData: identity.UserData{}},
			expected: false,
		},
		"no match": {
			lpa: &DonorProvidedDetails{
				Donor:                 Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{},
			},
			expected: false,
		},
		"none": {
			lpa:      &DonorProvidedDetails{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.DonorIdentityConfirmed())
		})
	}
}

func TestAttorneysSigningDeadline(t *testing.T) {
	donor := DonorProvidedDetails{
		SignedAt: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
	}

	expected := time.Date(2020, time.January, 30, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, donor.AttorneysAndCpSigningDeadline())
}

func TestUnder18ActorDetails(t *testing.T) {
	under18 := date.Today().AddDate(0, 0, -1)
	over18 := date.Today().AddDate(-18, 0, -1)
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	uid3 := actoruid.New()
	uid4 := actoruid.New()

	donor := DonorProvidedDetails{
		LpaID: "lpa-id",
		Attorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "a", LastName: "b", DateOfBirth: under18, UID: uid1},
			{FirstNames: "c", LastName: "d", DateOfBirth: over18, UID: uid2},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "e", LastName: "f", DateOfBirth: under18, UID: uid3},
			{FirstNames: "g", LastName: "h", DateOfBirth: over18, UID: uid4},
		}},
	}

	actors := donor.Under18ActorDetails()

	assert.Equal(t, []Under18ActorDetails{
		{FullName: "a b", DateOfBirth: under18, UID: uid1, Type: TypeAttorney},
		{FullName: "e f", DateOfBirth: under18, UID: uid3, Type: TypeReplacementAttorney},
	}, actors)
}

func TestAllAttorneysSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	otherLpaSignedAt := lpaSignedAt.Add(time.Minute)
	attorneySigned := lpaSignedAt.Add(time.Second)

	uid1 := actoruid.New()
	uid2 := actoruid.New()
	uid3 := actoruid.New()
	uid4 := actoruid.New()
	uid5 := actoruid.New()

	testcases := map[string]struct {
		lpa       *DonorProvidedDetails
		attorneys []*AttorneyProvidedDetails
		expected  bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid3}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid4, LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{UID: uid3, IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"need replacement attorney to sign": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{Attorneys: []Attorney{{UID: uid1}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid3}, {UID: uid5}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid3, IsReplacement: true},
				{UID: uid5, IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"all attorneys signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{UID: uid3}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid3, IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"more attorneys signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:  lpaSignedAt,
				Attorneys: Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{UID: uid4, LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"waiting for attorney to re-sign": {
			lpa: &DonorProvidedDetails{
				SignedAt:  lpaSignedAt,
				Attorneys: Attorneys{Attorneys: []Attorney{{UID: uid1}, {UID: uid2}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"trust corporations not signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.Yes,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: false,
		},
		"trust corporations signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					IsReplacement:            true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.AllAttorneysSigned(tc.attorneys))
		})
	}
}

func TestActorAddresses(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{Address: place.Address{Line1: "1"}},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{Address: place.Address{Line1: "2"}},
			{Address: place.Address{Line1: "3"}},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{Address: place.Address{Line1: "4"}},
			{Address: place.Address{Line1: "5"}},
		}},
		CertificateProvider: CertificateProvider{Address: place.Address{Line1: "6"}},
	}

	want := []place.Address{
		{Line1: "1"},
		{Line1: "6"},
		{Line1: "2"},
		{Line1: "3"},
		{Line1: "4"},
		{Line1: "5"},
	}

	assert.Equal(t, want, donor.ActorAddresses())
}

func TestActorAddressesActorWithNoAddressIgnored(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{FirstNames: "Donor", LastName: "Actor", Address: address},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "Attorney One", LastName: "Actor", Address: address},
			{FirstNames: "Attorney Two", LastName: "Actor"},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "Replacement Attorney One", LastName: "Actor"},
			{FirstNames: "Replacement Attorney Two", LastName: "Actor", Address: address},
		}},
		CertificateProvider: CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor"},
	}

	want := []place.Address{address}

	assert.Equal(t, want, donor.ActorAddresses())
}

func TestAllLayAttorneysFirstNames(t *testing.T) {
	donor := &DonorProvidedDetails{
		Attorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John", LastName: "Smith"},
				{FirstNames: "Barry", LastName: "Smith"},
			},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John2", LastName: "Smithe"},
				{FirstNames: "Barry2", LastName: "Smithe"},
			},
		},
	}

	assert.Equal(t, []string{"John", "Barry", "John2", "Barry2"}, donor.AllLayAttorneysFirstNames())
}

func TestAllLayAttorneysFullNames(t *testing.T) {
	donor := &DonorProvidedDetails{
		Attorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John", LastName: "Smith"},
				{FirstNames: "Barry", LastName: "Smith"},
			},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John2", LastName: "Smithe"},
				{FirstNames: "Barry2", LastName: "Smithe"},
			},
		},
	}

	assert.Equal(t, []string{"John Smith", "Barry Smith", "John2 Smithe", "Barry2 Smithe"}, donor.AllLayAttorneysFullNames())
}

func TestTrustCorporationOriginal(t *testing.T) {
	donor := &DonorProvidedDetails{
		Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "Corp"}},
		ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "Trust"}},
	}

	assert.Equal(t, []string{"Corp", "Trust"}, donor.TrustCorporationsNames())
}

func TestNamesChanged(t *testing.T) {
	testCases := map[string]*Donor{
		"FirstNames": {FirstNames: "d", LastName: "b", OtherNames: "c"},
		"LastName":   {FirstNames: "a", LastName: "d", OtherNames: "c"},
		"OtherNames": {FirstNames: "a", LastName: "b", OtherNames: "d"},
	}

	donor := &DonorProvidedDetails{Donor: Donor{FirstNames: "a", LastName: "b", OtherNames: "c"}}

	for name, updatedDonor := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.True(t, donor.NamesChanged(updatedDonor.FirstNames, updatedDonor.LastName, updatedDonor.OtherNames))
		})
	}

	assert.False(t, donor.NamesChanged("a", "b", "c"))
}

func TestIsOrganisationDonor(t *testing.T) {
	donor := &DonorProvidedDetails{SK: "ORGANISATION#123"}
	assert.True(t, donor.IsOrganisationDonor())

	donor.SK = ""

	assert.False(t, donor.IsOrganisationDonor())
}

func TestLpaProgressAsDonor(t *testing.T) {
	//dateOfBirth := date.Today()
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
				Tasks: DonorTasks{PayForLpa: PaymentTaskCompleted},
				//DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b", DateOfBirth: dateOfBirth},
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
			assert.Equal(t, tc.expectedProgress(), tc.donor.Progress(tc.certificateProvider, tc.attorneys, localizer))
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
		"initial state - with certificate provider name": {
			donor: &DonorProvidedDetails{
				SK:                  "ORGANISATION#123",
				Donor:               Donor{FirstNames: "a", LastName: "b"},
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
			assert.Equal(t, tc.expectedProgress(), tc.donor.Progress(tc.certificateProvider, tc.attorneys, localizer))
		})
	}
}
