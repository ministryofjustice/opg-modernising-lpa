package actor

import (
	"fmt"
	"strings"

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

type Attorneys []Attorney

func (as Attorneys) Get(id string) (Attorney, bool) {
	idx := slices.IndexFunc(as, func(a Attorney) bool { return a.ID == id })
	if idx == -1 {
		return Attorney{}, false
	}

	return as[idx], true
}

func (as Attorneys) Put(attorney Attorney) bool {
	idx := slices.IndexFunc(as, func(a Attorney) bool { return a.ID == attorney.ID })
	if idx == -1 {
		return false
	}

	as[idx] = attorney
	return true
}

func (as *Attorneys) Delete(attorney Attorney) bool {
	idx := slices.IndexFunc(*as, func(a Attorney) bool { return a.ID == attorney.ID })
	if idx == -1 {
		return false
	}

	*as = slices.Delete(*as, idx, idx+1)
	return true
}

func (as Attorneys) FullNames() string {
	names := make([]string, len(as))
	for i, a := range as {
		names[i] = fmt.Sprintf("%s %s", a.FirstNames, a.LastName)
	}

	return concatSentence(names)
}

func (as Attorneys) FirstNames() string {
	names := make([]string, len(as))
	for i, a := range as {
		names[i] = a.FirstNames
	}

	return concatSentence(names)
}

func concatSentence(list []string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		last := len(list) - 1
		return fmt.Sprintf("%s and %s", strings.Join(list[:last], ", "), list[last])
	}
}
