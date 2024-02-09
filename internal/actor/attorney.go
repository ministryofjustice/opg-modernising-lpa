package actor

import (
	"fmt"
	"slices"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// Attorney contains details about an attorney or replacement attorney, provided by the applicant
type Attorney struct {
	// UID for the actor
	UID UID
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
	// UID for the actor
	UID UID
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
	TrustCorporation TrustCorporation
	Attorneys        []Attorney
}

func (as Attorneys) Len() int {
	if as.TrustCorporation.Name == "" {
		return len(as.Attorneys)
	}

	return len(as.Attorneys) + 1
}

func (as Attorneys) Complete() bool {
	if as.TrustCorporation.Name != "" && as.TrustCorporation.Address.Line1 == "" {
		return false
	}

	for _, a := range as.Attorneys {
		if a.FirstNames == "" || (a.Address.Line1 == "" && a.Email == "") {
			return false
		}
	}

	return true
}

func (as Attorneys) Addresses() []place.Address {
	var addresses []place.Address

	if as.TrustCorporation.Address.String() != "" {
		addresses = append(addresses, as.TrustCorporation.Address)
	}

	for _, attorney := range as.Attorneys {
		if attorney.Address.String() != "" && !slices.Contains(addresses, attorney.Address) {
			addresses = append(addresses, attorney.Address)
		}
	}

	return addresses
}

func (as Attorneys) Get(uid UID) (Attorney, bool) {
	idx := as.Index(uid)
	if idx == -1 {
		return Attorney{}, false
	}

	return as.Attorneys[idx], true
}

func (as *Attorneys) Put(attorney Attorney) {
	idx := as.Index(attorney.UID)
	if idx == -1 {
		as.Attorneys = append(as.Attorneys, attorney)
	} else {
		as.Attorneys[idx] = attorney
	}
}

func (as *Attorneys) Delete(attorney Attorney) bool {
	idx := as.Index(attorney.UID)
	if idx == -1 {
		return false
	}

	as.Attorneys = slices.Delete(as.Attorneys, idx, idx+1)
	return true
}

func (as *Attorneys) Index(uid UID) int {
	return slices.IndexFunc(as.Attorneys, func(a Attorney) bool { return a.UID == uid })
}

func (as Attorneys) FullNames() []string {
	var names []string

	if as.TrustCorporation.Name != "" {
		names = append(names, as.TrustCorporation.Name)
	}

	for _, a := range as.Attorneys {
		names = append(names, fmt.Sprintf("%s %s", a.FirstNames, a.LastName))
	}

	return names
}

func (as Attorneys) FirstNames() []string {
	var names []string

	if as.TrustCorporation.Name != "" {
		names = append(names, as.TrustCorporation.Name)
	}

	for _, a := range as.Attorneys {
		names = append(names, a.FirstNames)
	}

	return names
}
