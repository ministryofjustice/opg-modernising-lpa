package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

//go:generate enumerator -type CorrespondentShare -linecomment -trimprefix -empty -bits
type CorrespondentShare uint8

const (
	CorrespondentShareAttorneys CorrespondentShare = 2 << iota
	CorrespondentShareCertificateProvider
)

type Correspondent struct {
	FirstNames   string
	LastName     string
	Email        string
	Organisation string
	Telephone    string
	WantAddress  form.YesNo
	Address      place.Address
	Share        CorrespondentShare
}

func (c Correspondent) FullName() string {
	return c.FirstNames + " " + c.LastName
}
