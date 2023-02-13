package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type taskListData struct {
	App      page.AppData
	Errors   validation.List
	Lpa      *page.Lpa
	Sections []taskListSection
}

type taskListItem struct {
	Name  string
	Path  string
	State page.TaskState
	Count int
}

type taskListSection struct {
	Heading string
	Items   []taskListItem
}

func TaskList(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
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
							Path:  page.Paths.YourDetails,
							State: lpa.Tasks.YourDetails,
						},
						{
							Name:  "chooseYourAttorneys",
							Path:  page.Paths.ChooseAttorneys,
							State: lpa.Tasks.ChooseAttorneys,
							Count: len(lpa.Attorneys),
						},
						{
							Name:  "chooseYourReplacementAttorneys",
							Path:  page.Paths.DoYouWantReplacementAttorneys,
							State: lpa.Tasks.ChooseReplacementAttorneys,
							Count: len(lpa.ReplacementAttorneys),
						},
						{
							Name:  "chooseWhenTheLpaCanBeUsed",
							Path:  page.Paths.WhenCanTheLpaBeUsed,
							State: lpa.Tasks.WhenCanTheLpaBeUsed,
						},
						{
							Name:  "addRestrictionsToTheLpa",
							Path:  page.Paths.Restrictions,
							State: lpa.Tasks.Restrictions,
						},
						{
							Name:  "chooseYourCertificateProvider",
							Path:  page.Paths.WhoDoYouWantToBeCertificateProviderGuidance,
							State: lpa.Tasks.CertificateProvider,
						},
						{
							Name:  "peopleToNotify",
							Path:  page.Paths.DoYouWantToNotifyPeople,
							State: lpa.Tasks.PeopleToNotify,
							Count: len(lpa.PeopleToNotify),
						},
						{
							Name:  "checkAndSendToYourCertificateProvider",
							Path:  page.Paths.CheckYourLpa,
							State: lpa.Tasks.CheckYourLpa,
						},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{
							Name:  "payForTheLpa",
							Path:  page.Paths.AboutPayment,
							State: lpa.Tasks.PayForLpa,
						},
					},
				},
				{
					Heading: "confirmYourIdentityAndSign",
					Items: []taskListItem{
						{
							Name:  "confirmYourIdentityAndSign",
							Path:  page.Paths.HowToConfirmYourIdentityAndSign,
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
