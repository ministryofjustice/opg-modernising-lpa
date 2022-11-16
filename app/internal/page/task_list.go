package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
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

		var attorneyAddressCompleted bool

		for _, a := range lpa.Attorneys {
			if a.Address.Line1 == "" {
				attorneyAddressCompleted = false
				break
			}

			attorneyAddressCompleted = true
		}

		var attorneyDetailsCompleted bool

		for _, a := range lpa.Attorneys {
			if a.FirstNames == "" {
				attorneyDetailsCompleted = false
				break
			}

			attorneyDetailsCompleted = true
		}

		data := &taskListData{
			App: appData,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{
							Name:       "provideDonorDetails",
							Path:       yourDetailsPath,
							Completed:  lpa.You.Address.Line1 != "",
							InProgress: lpa.You.FirstNames != "",
						},
						{
							Name:       "chooseYourAttorneys",
							Path:       chooseAttorneysPath,
							Completed:  attorneyAddressCompleted && attorneyDetailsCompleted,
							InProgress: attorneyDetailsCompleted && !attorneyAddressCompleted,
							Count:      len(lpa.Attorneys),
						},
						{
							Name:       "chooseYourReplacementAttorneys",
							Path:       wantReplacementAttorneysPath,
							Completed:  lpa.ReplacementAttorneysTaskComplete(),
							InProgress: !lpa.ReplacementAttorneysTaskComplete(),
							Count:      len(lpa.ReplacementAttorneys),
						},
						{
							Name:       "chooseWhenTheLpaCanBeUsed",
							Path:       whenCanTheLpaBeUsedPath,
							Completed:  lpa.Tasks.WhenCanTheLpaBeUsed == TaskCompleted,
							InProgress: lpa.Tasks.WhenCanTheLpaBeUsed == TaskInProgress,
						},
						{
							Name:       "addRestrictionsToTheLpa",
							Path:       restrictionsPath,
							Completed:  lpa.Tasks.Restrictions == TaskCompleted,
							InProgress: lpa.Tasks.Restrictions == TaskInProgress,
						},
						{
							Name:       "chooseYourCertificateProvider",
							Path:       whoDoYouWantToBeCertificateProviderGuidancePath,
							Completed:  lpa.Tasks.CertificateProvider == TaskCompleted,
							InProgress: lpa.Tasks.CertificateProvider == TaskInProgress,
						},
						{
							Name:       "checkAndSendToYourCertificateProvider",
							Path:       checkYourLpaPath,
							Completed:  lpa.Tasks.CheckYourLpa == TaskCompleted,
							InProgress: lpa.Tasks.CheckYourLpa == TaskInProgress,
						},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{
							Name:       "payForTheLpa",
							Path:       aboutPaymentPath,
							Completed:  lpa.Tasks.PayForLpa == TaskCompleted,
							InProgress: lpa.Tasks.PayForLpa == TaskInProgress,
						},
					},
				},
				{
					Heading: "confirmYourIdentityAndSign",
					Items: []taskListItem{
						{
							Name:       "confirmYourIdentityAndSign",
							Path:       selectYourIdentityOptionsPath,
							Completed:  lpa.Tasks.ConfirmYourIdentityAndSign == TaskCompleted,
							InProgress: lpa.Tasks.ConfirmYourIdentityAndSign == TaskInProgress,
						},
					},
				},
				{
					Heading: "registerTheLpa",
					Items: []taskListItem{
						{Name: "registerTheLpa", Disabled: true},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
