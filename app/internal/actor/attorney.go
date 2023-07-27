package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
	"golang.org/x/exp/slices"
)

// Attorney contains details about an attorney or replacement attorney, provided by the applicant
type Attorney struct {
	// Identifies the attorney being edited
	ID string
	// First names of the attorney
	FirstNames string
	// Last name of the attorney
	LastName string
	// Email of the attorney
	Email string
	// Date of birth of the attorney
	DateOfBirth date.Date
	// Address of the attorney
	Address place.Address
}

func (a Attorney) FullName() string {
	return fmt.Sprintf("%s %s", a.FirstNames, a.LastName)
}

// TrustCorporation contains details about a trust corporation, provided by the applicant
type TrustCorporation struct {
	// Name of the company
	Name string
	// CompanyNumber as registered by Companies House
	CompanyNumber string
	// Email to contact the company
	Email string
	// Address of the company
	Address place.Address
}

type Attorneys struct {
	trustCorporation *TrustCorporation
	attorneys        []Attorney
}

func NewAttorneys(tc *TrustCorporation, as []Attorney) Attorneys {
	return Attorneys{
		trustCorporation: tc,
		attorneys:        as,
	}
}

func (as Attorneys) Len() int {
	if as.trustCorporation == nil {
		return len(as.attorneys)
	}

	return len(as.attorneys) + 1
}

func (as Attorneys) Complete() bool {
	if tc, ok := as.TrustCorporation(); ok {
		if tc.Name == "" || tc.Address.Line1 == "" {
			return false
		}
	}

	for _, a := range as.attorneys {
		if a.FirstNames == "" || (a.Address.Line1 == "" && a.Email == "") {
			return false
		}
	}

	return true
}

func (as Attorneys) Addresses() []place.Address {
	var addresses []place.Address

	if tc, ok := as.TrustCorporation(); ok {
		if tc.Address.String() != "" && !slices.Contains(addresses, tc.Address) {
			addresses = append(addresses, tc.Address)
		}
	}

	for _, attorney := range as.Attorneys() {
		if attorney.Address.String() != "" && !slices.Contains(addresses, attorney.Address) {
			addresses = append(addresses, attorney.Address)
		}
	}

	return addresses
}

func (as Attorneys) TrustCorporation() (TrustCorporation, bool) {
	if as.trustCorporation == nil {
		return TrustCorporation{}, false
	}

	return *as.trustCorporation, true
}

func (as *Attorneys) SetTrustCorporation(tc TrustCorporation) {
	as.trustCorporation = &tc
}

func (as Attorneys) Attorneys() []Attorney {
	return as.attorneys
}

func (as Attorneys) Get(id string) (Attorney, bool) {
	idx := slices.IndexFunc(as.attorneys, func(a Attorney) bool { return a.ID == id })
	if idx == -1 {
		return Attorney{}, false
	}

	return as.attorneys[idx], true
}

func (as *Attorneys) Put(attorney Attorney) {
	idx := slices.IndexFunc(as.attorneys, func(a Attorney) bool { return a.ID == attorney.ID })
	if idx == -1 {
		as.attorneys = append(as.attorneys, attorney)
	} else {
		as.attorneys[idx] = attorney
	}
}

func (as *Attorneys) Delete(attorney Attorney) bool {
	idx := slices.IndexFunc(as.attorneys, func(a Attorney) bool { return a.ID == attorney.ID })
	if idx == -1 {
		return false
	}

	as.attorneys = slices.Delete(as.attorneys, idx, idx+1)
	return true
}

func (as Attorneys) FullNames() []string {
	names := make([]string, len(as.attorneys))
	for i, a := range as.attorneys {
		names[i] = fmt.Sprintf("%s %s", a.FirstNames, a.LastName)
	}

	return names
}

func (as Attorneys) FirstNames() []string {
	names := make([]string, len(as.attorneys))
	for i, a := range as.attorneys {
		names[i] = a.FirstNames
	}

	return names
}
