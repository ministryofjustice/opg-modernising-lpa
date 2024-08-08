package lpadata

import (
	"fmt"
	"slices"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
)

type Attorneys struct {
	Attorneys        []Attorney
	TrustCorporation TrustCorporation
}

func (a Attorneys) Len() int {
	if a.TrustCorporation.Name != "" {
		return 1 + len(a.Attorneys)
	}

	return len(a.Attorneys)
}

func (a Attorneys) Index(uid actoruid.UID) int {
	return slices.IndexFunc(a.Attorneys, func(a Attorney) bool { return a.UID == uid })
}

func (a Attorneys) Get(uid actoruid.UID) (Attorney, bool) {
	idx := a.Index(uid)
	if idx == -1 {
		return Attorney{}, false
	}

	return a.Attorneys[idx], true
}

func (a Attorneys) FullNames() []string {
	var names []string

	if a.TrustCorporation.Name != "" {
		names = append(names, a.TrustCorporation.Name)
	}

	for _, a := range a.Attorneys {
		names = append(names, fmt.Sprintf("%s %s", a.FirstNames, a.LastName))
	}

	return names
}
