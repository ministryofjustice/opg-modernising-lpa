package donordata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/task"

type Tasks struct {
	YourDetails                task.State
	ChooseAttorneys            task.State
	ChooseReplacementAttorneys task.State
	WhenCanTheLpaBeUsed        task.State // property and affairs only
	LifeSustainingTreatment    task.State // personal welfare only
	Restrictions               task.State
	CertificateProvider        task.State
	PeopleToNotify             task.State
	AddCorrespondent           task.State
	ChooseYourSignatory        task.State // if .Donor.CanSign.IsNo only
	CheckYourLpa               task.State
	PayForLpa                  task.PaymentState
	ConfirmYourIdentityAndSign task.IdentityState
}
