package actor

import (
	"fmt"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
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
	makeDonor := func(version uint8, hash uint64) *DonorProvidedDetails {
		return &DonorProvidedDetails{
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
	const modified uint64 = 0x6cc76d93a3e8e7d8

	// DO NOT change these initial hash values. If a field has been added/removed
	// you will need to handle the version gracefully by modifying
	// (*DonorProvidedDetails).HashInclude and adding another testcase for the new
	// version.
	testcases := map[uint8]uint64{
		0: 0x228526f2932e1744,
		1: 0xb4184bcfa870e40a,
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
			assert.Equal(t, uint8(1), donor.HashVersion)
		})
	}
}

func TestGenerateHashVersionTooHigh(t *testing.T) {
	donor := &DonorProvidedDetails{
		HashVersion: 2,
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
	makeDonor := func(version uint8, hash uint64) *DonorProvidedDetails {
		return &DonorProvidedDetails{
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
	const modified uint64 = 0x91100db15d995176

	// DO NOT change these initial hash values. If a field has been added/removed
	// you will need to handle the version gracefully by modifying
	// toCheck.HashInclude and adding another testcase for the new version.
	testcases := map[uint8]uint64{
		0: 0x61d5afc9bc5a9de7,
		1: 0x88b192c5eb912669,
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
			assert.Equal(t, uint8(1), donor.CheckedHashVersion)
		})
	}
}

func TestGenerateCheckedHashVersionTooHigh(t *testing.T) {
	donor := &DonorProvidedDetails{
		CheckedHashVersion: 2,
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
