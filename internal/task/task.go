package task

type DonorTasks struct {
	YourDetails                State
	ChooseAttorneys            State
	ChooseReplacementAttorneys State
	WhenCanTheLpaBeUsed        State // property and affairs only
	LifeSustainingTreatment    State // personal welfare only
	Restrictions               State
	CertificateProvider        State
	PeopleToNotify             State
	AddCorrespondent           State
	ChooseYourSignatory        State // if .Donor.CanSign.IsNo only
	CheckYourLpa               State
	PayForLpa                  PaymentState
	ConfirmYourIdentityAndSign IdentityState
}
