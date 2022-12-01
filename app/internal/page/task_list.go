package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

const (
	FillInLpaSection                      = "fillInTheLpa"
	ProvideYourDetailsTask                = "provideYourDetails"
	ChooseYourAttorneysTask               = "chooseYourAttorneys"
	ChooseYourReplacementAttorneysTask    = "chooseYourReplacementAttorneys"
	ChooseWhenTheLpaCanBeUsedTask         = "chooseWhenTheLpaCanBeUsed"
	AddRestrictionsToLpaTask              = "addRestrictionsToTheLpa"
	ChooseCertificateProviderTask         = "chooseYourCertificateProvider"
	CheckAndSendToCertificateProviderTask = "checkAndSendToYourCertificateProvider"
	PayForLpaSection                      = "payForTheLpa"
	PayForTheLpaTask                      = "payForTheLpa"
	ConfirmYourIdentityAndSignSection     = "confirmYourIdentityAndSign"
	ConfirmYourIdentityAndSignTask        = "confirmYourIdentityAndSign"
	RegisterTheLpaSection                 = "registerTheLpa"
	RegisterTheLpaTask                    = "registerTheLpa"
)

type taskListData struct {
	App      AppData
	Errors   map[string]string
	Sections []taskListSection
}

type taskListItem struct {
	Name       string
	Path       string
	Completed  bool
	InProgress bool
	Disabled   bool
	Count      int
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &taskListData{
			App: appData,
			Sections: []taskListSection{
				{
					Heading: FillInLpaSection,
					Items: []taskListItem{
						{
							Name:       ProvideYourDetailsTask,
							Path:       yourDetailsPath,
							Completed:  lpa.You.Address.Line1 != "",
							InProgress: lpa.You.FirstNames != "",
						},
						{
							Name:       ChooseYourAttorneysTask,
							Path:       chooseAttorneysPath,
							Completed:  lpa.AttorneysTaskComplete(),
							InProgress: len(lpa.Attorneys) > 0 && !lpa.AttorneysTaskComplete(),
							Count:      len(lpa.Attorneys),
						},
						{
							Name:       ChooseYourReplacementAttorneysTask,
							Path:       wantReplacementAttorneysPath,
							Completed:  lpa.ReplacementAttorneysTaskComplete(),
							InProgress: len(lpa.ReplacementAttorneys) > 0 && !lpa.ReplacementAttorneysTaskComplete(),
							Count:      len(lpa.ReplacementAttorneys),
						},
						{
							Name:       ChooseWhenTheLpaCanBeUsedTask,
							Path:       whenCanTheLpaBeUsedPath,
							Completed:  lpa.Tasks.WhenCanTheLpaBeUsed == TaskCompleted,
							InProgress: lpa.Tasks.WhenCanTheLpaBeUsed == TaskInProgress,
						},
						{
							Name:       AddRestrictionsToLpaTask,
							Path:       restrictionsPath,
							Completed:  lpa.Tasks.Restrictions == TaskCompleted,
							InProgress: lpa.Tasks.Restrictions == TaskInProgress,
						},
						{
							Name:       ChooseCertificateProviderTask,
							Path:       whoDoYouWantToBeCertificateProviderGuidancePath,
							Completed:  lpa.Tasks.CertificateProvider == TaskCompleted,
							InProgress: lpa.Tasks.CertificateProvider == TaskInProgress,
						},
						{
							Name:       CheckAndSendToCertificateProviderTask,
							Path:       checkYourLpaPath,
							Completed:  lpa.Tasks.CheckYourLpa == TaskCompleted,
							InProgress: lpa.Tasks.CheckYourLpa == TaskInProgress,
						},
					},
				},
				{
					Heading: PayForLpaSection,
					Items: []taskListItem{
						{
							Name:       PayForTheLpaTask,
							Path:       aboutPaymentPath,
							Completed:  lpa.Tasks.PayForLpa == TaskCompleted,
							InProgress: lpa.Tasks.PayForLpa == TaskInProgress,
						},
					},
				},
				{
					Heading: ConfirmYourIdentityAndSignSection,
					Items: []taskListItem{
						{
							Name:       ConfirmYourIdentityAndSignTask,
							Path:       selectYourIdentityOptionsPath,
							Completed:  lpa.Tasks.ConfirmYourIdentityAndSign == TaskCompleted,
							InProgress: lpa.Tasks.ConfirmYourIdentityAndSign == TaskInProgress,
						},
					},
				},
				{
					Heading: RegisterTheLpaSection,
					Items: []taskListItem{
						{Name: RegisterTheLpaTask, Disabled: true},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
