package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type taskListData struct {
	App      AppData
	Errors   map[string]string
	Lpa      *Lpa
	Sections []taskListSection
}

type taskListItem struct {
	Name  string
	Path  string
	State TaskState
	Count int
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
			Lpa: lpa,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{
							Name:  "provideYourDetails",
							Path:  Paths.YourDetails,
							State: lpa.Tasks.YourDetails,
						},
						{
							Name:  "chooseYourAttorneys",
							Path:  Paths.ChooseAttorneys,
							State: lpa.Tasks.ChooseAttorneys,
							Count: len(lpa.Attorneys),
						},
						{
							Name:  "chooseYourReplacementAttorneys",
							Path:  Paths.DoYouWantReplacementAttorneys,
							State: lpa.Tasks.ChooseReplacementAttorneys,
							Count: len(lpa.ReplacementAttorneys),
						},
						{
							Name:  "chooseWhenTheLpaCanBeUsed",
							Path:  Paths.WhenCanTheLpaBeUsed,
							State: lpa.Tasks.WhenCanTheLpaBeUsed,
						},
						{
							Name:  "addRestrictionsToTheLpa",
							Path:  Paths.Restrictions,
							State: lpa.Tasks.Restrictions,
						},
						{
							Name:  "chooseYourCertificateProvider",
							Path:  Paths.WhoDoYouWantToBeCertificateProviderGuidance,
							State: lpa.Tasks.CertificateProvider,
						},
						{
							Name:  "peopleToNotify",
							Path:  Paths.DoYouWantToNotifyPeople,
							State: lpa.Tasks.PeopleToNotify,
							Count: len(lpa.PeopleToNotify),
						},
						{
							Name:  "checkAndSendToYourCertificateProvider",
							Path:  Paths.CheckYourLpa,
							State: lpa.Tasks.CheckYourLpa,
						},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{
							Name:  "payForTheLpa",
							Path:  Paths.AboutPayment,
							State: lpa.Tasks.PayForLpa,
						},
					},
				},
				{
					Heading: "confirmYourIdentityAndSign",
					Items: []taskListItem{
						{
							Name:  "confirmYourIdentityAndSign",
							Path:  Paths.HowToConfirmYourIdentityAndSign,
							State: lpa.Tasks.ConfirmYourIdentityAndSign,
						},
					},
				},
				{
					Heading: "registerTheLpa",
					Items: []taskListItem{
						{
							Name: "registerTheLpa",
						},
					},
				},
			},
		}

		return tmpl(w, data)
	}
}
