package donordata

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
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

func TestProvidedCanChange(t *testing.T) {
	assert.True(t, (&Provided{}).CanChange())
	assert.False(t, (&Provided{SignedAt: time.Now()}).CanChange())
}

func TestProvidedCanChangePersonalDetails(t *testing.T) {
	testcases := map[string]struct {
		provided  Provided
		canChange bool
	}{
		"no personal details": {
			provided:  Provided{},
			canChange: true,
		},
		"signed": {
			provided: Provided{
				SignedAt: time.Now(),
			},
		},
		"identity confirmed": {
			provided: Provided{
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			},
		},
		"vouch in progress": {
			provided: Provided{
				VoucherInvitedAt: testNow,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.canChange, tc.provided.CanChangePersonalDetails())
		})
	}
}

func TestGenerateHash(t *testing.T) {
	makeDonor := func(version uint8, hash uint64) *Provided {
		return &Provided{
			HashVersion: version,
			Hash:        hash,
			Attorneys: Attorneys{
				Attorneys: []Attorney{
					{DateOfBirth: date.New("2000", "1", "2")},
				},
			},
		}
	}

	// DO change this value to match the updates
	const modified uint64 = 0x88bad809bb9c1575

	// DO NOT change these initial hash values. If a field has been added/removed
	// you will need to handle the version gracefully by modifying
	// (*Provided).HashInclude and adding another testcase for the new
	// version.
	testcases := map[uint8]uint64{
		0: 0xb23345a21655d556,
	}

	for version, initial := range testcases {
		t.Run(fmt.Sprintf("Version%d", version), func(t *testing.T) {
			donor := makeDonor(version, initial)
			hash, _ := donor.generateHash()

			assert.Equal(t, donor.Hash, hash)
			assert.False(t, donor.HashChanged())

			donor.Attorneys.Attorneys[0].DateOfBirth = date.New("2001", "1", "2")
			assert.True(t, donor.HashChanged())

			err := donor.UpdateHash()
			assert.Nil(t, err)
			assert.Equal(t, modified, donor.Hash)
			assert.Equal(t, uint8(currentHashVersion), donor.HashVersion)
		})
	}
}

func TestGenerateHashVersionTooHigh(t *testing.T) {
	donor := &Provided{
		HashVersion: currentHashVersion + 1,
		Attorneys: Attorneys{
			Attorneys: []Attorney{
				{DateOfBirth: date.New("2000", "1", "2")},
			},
		},
	}

	_, err := donor.generateHash()
	assert.Error(t, err)
}

func TestGenerateCheckedHash(t *testing.T) {
	makeDonor := func(version uint8, hash uint64) *Provided {
		return &Provided{
			CheckedHashVersion: version,
			CheckedHash:        hash,
			Attorneys: Attorneys{
				Attorneys: []Attorney{
					{DateOfBirth: date.New("2000", "1", "2")},
				},
			},
		}
	}

	// DO change this value to match the updates
	const modified uint64 = 0x70d44213a2b8cce3

	// DO NOT change these initial hash values. If a field has been added/removed
	// you will need to handle the version gracefully by modifying
	// toCheck.HashInclude and adding another testcase for the new version.
	testcases := map[uint8]uint64{
		0: 0xa9fccd755b8d87b7,
	}

	for version, initial := range testcases {
		t.Run(fmt.Sprintf("Version%d", version), func(t *testing.T) {
			donor := makeDonor(version, initial)
			hash, _ := donor.generateCheckedHash()

			assert.Equal(t, donor.CheckedHash, hash)
			assert.False(t, donor.CheckedHashChanged())

			donor.Attorneys.Attorneys[0].DateOfBirth = date.New("2001", "1", "2")
			assert.True(t, donor.CheckedHashChanged())

			err := donor.UpdateCheckedHash()
			assert.Nil(t, err)
			assert.Equal(t, modified, donor.CheckedHash)
			assert.Equal(t, uint8(currentCheckedHashVersion), donor.CheckedHashVersion)
		})
	}
}

func TestGenerateCheckedHashVersionTooHigh(t *testing.T) {
	donor := &Provided{
		CheckedHashVersion: currentCheckedHashVersion + 1,
		Attorneys: Attorneys{
			Attorneys: []Attorney{
				{DateOfBirth: date.New("2000", "1", "2")},
			},
		},
	}

	_, err := donor.generateCheckedHash()
	assert.Error(t, err)
}

func TestIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Provided
		expected bool
	}{
		"confirmed": {
			lpa: &Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.StatusConfirmed, DateOfBirth: date.New("2000", "1", "1")},
				Donor: Donor{
					FirstNames:  "a",
					LastName:    "b",
					DateOfBirth: date.New("2000", "1", "1"),
				},
			},
			expected: true,
		},
		"failed": {
			lpa: &Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.StatusFailed, DateOfBirth: date.New("2000", "1", "1")},
				Donor: Donor{
					FirstNames:  "a",
					LastName:    "b",
					DateOfBirth: date.New("2000", "1", "1"),
				},
			},
			expected: false,
		},
		"name does not match": {
			lpa: &Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.StatusConfirmed, DateOfBirth: date.New("2000", "1", "1")},
				Donor: Donor{
					FirstNames:  "a",
					LastName:    "c",
					DateOfBirth: date.New("2000", "1", "1"),
				},
			},
			expected: false,
		},
		"dob does not match": {
			lpa: &Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.StatusConfirmed, DateOfBirth: date.New("2000", "1", "1")},
				Donor: Donor{
					FirstNames:  "a",
					LastName:    "b",
					DateOfBirth: date.New("2000", "1", "2"),
				},
			},
			expected: false,
		},
		"insufficient evidence": {
			lpa: &Provided{
				IdentityUserData: identity.UserData{FirstNames: "a", LastName: "b", Status: identity.StatusInsufficientEvidence, DateOfBirth: date.New("2000", "1", "1")},
				Donor: Donor{
					FirstNames:  "a",
					LastName:    "b",
					DateOfBirth: date.New("2000", "1", "1"),
				},
			},
			expected: false,
		},
		"none": {
			lpa:      &Provided{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.DonorIdentityConfirmed())
		})
	}
}

func TestSigningDeadline(t *testing.T) {
	donor := Provided{
		SignedAt: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
	}

	expected := time.Date(2020, time.January, 30, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, donor.SigningDeadline())

	donor.RegisteringWithCourtOfProtection = true
	expected = time.Date(2020, time.May, 16, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, donor.SigningDeadline())
}

func TestCourtOfProtectionSubmissionDeadline(t *testing.T) {
	donor := Provided{
		SignedAt: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
	}

	expected := time.Date(2020, time.July, 2, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, donor.CourtOfProtectionSubmissionDeadline())
}

func TestUnder18ActorDetails(t *testing.T) {
	under18 := date.Today().AddDate(0, 0, -1)
	over18 := date.Today().AddDate(-18, 0, -1)
	uid1 := actoruid.New()
	uid2 := actoruid.New()
	uid3 := actoruid.New()
	uid4 := actoruid.New()

	donor := Provided{
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
		{FullName: "a b", DateOfBirth: under18, UID: uid1, Type: actor.TypeAttorney},
		{FullName: "e f", DateOfBirth: under18, UID: uid3, Type: actor.TypeReplacementAttorney},
	}, actors)
}

func TestProvidedCorrespondentEmail(t *testing.T) {
	lpa := &Provided{
		Donor: Donor{Email: "donor"},
	}
	assert.Equal(t, "donor", lpa.CorrespondentEmail())
}

func TestProvidedCorrespondentEmailWhenCorrespondentProvided(t *testing.T) {
	lpa := &Provided{
		Donor:         Donor{Email: "donor"},
		Correspondent: Correspondent{Email: "correspondent"},
	}
	assert.Equal(t, "correspondent", lpa.CorrespondentEmail())
}

func TestActorAddresses(t *testing.T) {
	donor := &Provided{
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
	donor := &Provided{
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
	donor := &Provided{
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
	donor := &Provided{
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
	donor := &Provided{
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

	donor := &Provided{Donor: Donor{FirstNames: "a", LastName: "b", OtherNames: "c"}}

	for name, updatedDonor := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.True(t, donor.NamesChanged(updatedDonor.FirstNames, updatedDonor.LastName, updatedDonor.OtherNames))
		})
	}

	assert.False(t, donor.NamesChanged("a", "b", "c"))
}

func TestProvidedActors(t *testing.T) {
	lpa := &Provided{
		Donor: Donor{
			UID:        actoruid.New(),
			FirstNames: "Sam",
			LastName:   "Smith",
		},
		CertificateProvider: CertificateProvider{
			UID:        actoruid.New(),
			FirstNames: "Charlie",
			LastName:   "Cooper",
		},
		Attorneys: Attorneys{
			Attorneys: []Attorney{{
				UID:        actoruid.New(),
				FirstNames: "Alan",
				LastName:   "Attorney",
			}, {
				UID:        actoruid.New(),
				FirstNames: "Angela",
				LastName:   "Attorney",
			}},
			TrustCorporation: TrustCorporation{Name: "Trusty"},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys: []Attorney{{
				UID:        actoruid.New(),
				FirstNames: "Richard",
				LastName:   "Replacement",
			}, {
				UID:        actoruid.New(),
				FirstNames: "Rachel",
				LastName:   "Replacement",
			}},
			TrustCorporation: TrustCorporation{Name: "Untrusty"},
		},
		PeopleToNotify: []PersonToNotify{{
			UID:        actoruid.New(),
			FirstNames: "Peter",
			LastName:   "Person",
		}},
		AuthorisedSignatory: AuthorisedSignatory{
			FirstNames: "Arthur",
			LastName:   "Signor",
		},
		IndependentWitness: IndependentWitness{
			FirstNames: "Independent",
			LastName:   "Wit",
		},
		Correspondent: Correspondent{FirstNames: "Nope"},
		Voucher:       Voucher{FirstNames: "Nada"},
	}

	actors := slices.Collect(lpa.Actors())

	assert.Equal(t, []actor.Actor{{
		Type:       actor.TypeDonor,
		UID:        lpa.Donor.UID,
		FirstNames: "Sam",
		LastName:   "Smith",
	}, {
		Type:       actor.TypeCertificateProvider,
		UID:        lpa.CertificateProvider.UID,
		FirstNames: "Charlie",
		LastName:   "Cooper",
	}, {
		Type:       actor.TypeAttorney,
		UID:        lpa.Attorneys.Attorneys[0].UID,
		FirstNames: "Alan",
		LastName:   "Attorney",
	}, {
		Type:       actor.TypeAttorney,
		UID:        lpa.Attorneys.Attorneys[1].UID,
		FirstNames: "Angela",
		LastName:   "Attorney",
	}, {
		Type:       actor.TypeReplacementAttorney,
		UID:        lpa.ReplacementAttorneys.Attorneys[0].UID,
		FirstNames: "Richard",
		LastName:   "Replacement",
	}, {
		Type:       actor.TypeReplacementAttorney,
		UID:        lpa.ReplacementAttorneys.Attorneys[1].UID,
		FirstNames: "Rachel",
		LastName:   "Replacement",
	}, {
		Type:       actor.TypePersonToNotify,
		UID:        lpa.PeopleToNotify[0].UID,
		FirstNames: "Peter",
		LastName:   "Person",
	}, {
		Type:       actor.TypeAuthorisedSignatory,
		FirstNames: "Arthur",
		LastName:   "Signor",
	}, {
		Type:       actor.TypeIndependentWitness,
		FirstNames: "Independent",
		LastName:   "Wit",
	}}, actors)
}

func TestProvidedCanHaveVoucher(t *testing.T) {
	provided := Provided{}
	assert.True(t, provided.CanHaveVoucher())

	provided.FailedVouchAttempts++
	assert.True(t, provided.CanHaveVoucher())

	provided.FailedVouchAttempts++
	assert.False(t, provided.CanHaveVoucher())

	provided.FailedVouchAttempts++
	assert.False(t, provided.CanHaveVoucher())
}

func TestProvidedUpdateDecisions(t *testing.T) {
	t.Run("no attorneys means no decisions", func(t *testing.T) {
		actual := &Provided{
			AttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{}, actual)
	})

	t.Run("one attorney means no decisions", func(t *testing.T) {
		actual := &Provided{
			Attorneys:         Attorneys{Attorneys: []Attorney{{}}},
			AttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys: Attorneys{Attorneys: []Attorney{{}}},
		}, actual)
	})

	t.Run("many attorneys jointly means no details", func(t *testing.T) {
		actual := &Provided{
			Attorneys:         Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly, Details: "what"},
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:         Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}, actual)
	})

	t.Run("many attorneys jointly for some means no change", func(t *testing.T) {
		actual := &Provided{
			Attorneys:         Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "what"},
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:         Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "what"},
		}, actual)
	})

	t.Run("no replacement attorneys means no decisions", func(t *testing.T) {
		actual := &Provided{
			ReplacementAttorneys:         Attorneys{Attorneys: []Attorney{{}}},
			ReplacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{}}},
		}, actual)
	})

	t.Run("one replacement attorney means no decisions", func(t *testing.T) {
		actual := &Provided{
			ReplacementAttorneys:         Attorneys{Attorneys: []Attorney{{}}},
			ReplacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{}}},
		}, actual)
	})

	t.Run("many replacement attorneys and one attorney means no details and no step in", func(t *testing.T) {
		actual := &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}}},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly, Details: "hey"},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:                    Attorneys{Attorneys: []Attorney{{}}},
			ReplacementAttorneys:         Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}, actual)
	})

	t.Run("many replacement attorneys and jointly attorney means no details and no step in", func(t *testing.T) {
		actual := &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.Jointly},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly, Details: "hey"},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:                    Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:            AttorneyDecisions{How: lpadata.Jointly},
			ReplacementAttorneys:         Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
		}, actual)
	})

	t.Run("many attorneys jointly and severally and many replacement attorneys one can act means step in but no decisions", func(t *testing.T) {
		actual := &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly, Details: "hey"},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
		}, actual)
	})

	t.Run("many attorneys jointly and severally and one replacement attorney all can act means step in but no decisions", func(t *testing.T) {
		actual := &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly, Details: "hey"},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}}},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}, actual)
	})

	t.Run("many attorneys jointly and severally and many replacement attorneys all can no longer act means step in and decisions", func(t *testing.T) {
		actual := &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly, Details: "hey"},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}, actual)
	})

	t.Run("many attorneys jointly for some means no step in and no decisions", func(t *testing.T) {
		actual := &Provided{
			Attorneys:                           Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:                   AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "hey"},
			ReplacementAttorneys:                Attorneys{Attorneys: []Attorney{{}, {}}},
			ReplacementAttorneyDecisions:        AttorneyDecisions{How: lpadata.Jointly},
			HowShouldReplacementAttorneysStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
		}
		actual.UpdateDecisions()

		assert.Equal(t, &Provided{
			Attorneys:            Attorneys{Attorneys: []Attorney{{}, {}}},
			AttorneyDecisions:    AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "hey"},
			ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{}, {}}},
		}, actual)
	})
}
