package lpadata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"

type Voucher struct {
	UID        actoruid.UID
	FirstNames string
	LastName   string
	Email      string
}
