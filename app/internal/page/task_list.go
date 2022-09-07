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
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
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
							Completed:  lpa.Attorney.Address.Line1 != "",
							InProgress: lpa.Attorney.FirstNames != "",
						},
						{
							Name:      "chooseYourReplacementAttorneys",
							Path:      wantReplacementAttorneysPath,
							Completed: lpa.WantReplacementAttorneys != "",
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
							InProgress: lpa.Tasks.WhoDoYouWantToBeCertificateProvider == TaskInProgress,
						},
						{Name: "checkAndSendToYourCertificateProvider"},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{Name: "payForTheLpa"},
					},
				},
				{
					Heading: "confirmYourIdentity",
					Items: []taskListItem{
						{Name: "confirmYourIdentity"},
					},
				},
				{
					Heading: "signAndRegisterTheLpa",
					Items: []taskListItem{
						{Name: "signTheLpa", Disabled: true},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
