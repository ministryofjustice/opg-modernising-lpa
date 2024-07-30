package actor

import "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"

type CanBeUsedWhen = donordata.CanBeUsedWhen

const (
	CanBeUsedWhenUnknown      = donordata.CanBeUsedWhenUnknown
	CanBeUsedWhenCapacityLost = donordata.CanBeUsedWhenCapacityLost
	CanBeUsedWhenHasCapacity  = donordata.CanBeUsedWhenHasCapacity
)
