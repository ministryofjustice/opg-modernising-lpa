package actor

import (
	"fmt"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"golang.org/x/exp/slices"
)

type Attorney struct {
	ID          string
	FirstNames  string
	LastName    string
	Email       string
	DateOfBirth date.Date
	Address     place.Address
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
