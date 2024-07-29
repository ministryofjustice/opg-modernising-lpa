package donordata

import "fmt"

// AuthorisedSignatory contains details of the person who will sign the LPA on the donor's behalf
type AuthorisedSignatory struct {
	FirstNames string
	LastName   string
}

func (s AuthorisedSignatory) FullName() string {
	return fmt.Sprintf("%s %s", s.FirstNames, s.LastName)
}
