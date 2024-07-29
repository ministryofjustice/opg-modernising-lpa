package actor

import "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"

type AttorneysAct = donordata.AttorneysAct

const (
	Jointly                          = donordata.Jointly
	JointlyAndSeverally              = donordata.JointlyAndSeverally
	JointlyForSomeSeverallyForOthers = donordata.JointlyForSomeSeverallyForOthers
)

type AttorneyDecisions = donordata.AttorneyDecisions

var MakeAttorneyDecisions = donordata.MakeAttorneyDecisions
