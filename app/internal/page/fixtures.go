package page

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"

	"golang.org/x/exp/slices"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

const TestEmail = "simulate-delivered@notifications.service.gov.uk"
const TestMobile = "07700900000"

var AttorneyNames = []string{
	"John",
	"Joan",
	"Johan",
	"Jilly",
	"James",
}

var ReplacementAttorneyNames = []string{
	"Jane",
	"Jorge",
	"Jackson",
	"Jacob",
	"Joshua",
}

var PeopleToNotifyNames = []string{
	"Joanna",
	"Jonathan",
	"Julian",
	"Jayden",
	"Juniper",
}

func MakePerson() actor.Donor {
	return actor.Donor{
		FirstNames: "Jose",
		LastName:   "Smith",
		Address: place.Address{
			Line1:      "1 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
		Email:       TestEmail,
		DateOfBirth: date.New("2000", "1", "2"),
	}
}

func MakeAttorney(firstNames string) actor.Attorney {
	return actor.Attorney{
		ID:          firstNames + "Smith",
		FirstNames:  firstNames,
		LastName:    "Smith",
		Email:       TestEmail,
		DateOfBirth: date.New("2000", "1", "2"),
		Address: place.Address{
			Line1:      "2 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
	}
}

func MakePersonToNotify(firstNames string) actor.PersonToNotify {
	return actor.PersonToNotify{
		ID:         firstNames + "Smith",
		FirstNames: firstNames,
		LastName:   "Smith",
		Email:      TestEmail,
		Address: place.Address{
			Line1:      "4 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
	}
}

func MakeCertificateProvider(firstNames string) actor.CertificateProvider {
	return actor.CertificateProvider{
		FirstNames:              firstNames,
		LastName:                "Smith",
		Email:                   TestEmail,
		Mobile:                  TestMobile,
		DateOfBirth:             date.New("1997", "1", "2"),
		Relationship:            "friend",
		RelationshipDescription: "",
		RelationshipLength:      "gte-2-years",
	}
}

func CompleteDonorDetails(lpa *Lpa) {
	lpa.Donor = MakePerson()
	lpa.WhoFor = "me"
	lpa.Type = "pfa"
	lpa.Tasks.YourDetails = TaskCompleted
}

func AddAttorneys(lpa *Lpa, count int) []string {
	if count > len(AttorneyNames) {
		count = len(AttorneyNames)
	}

	var firstNames []string
	for _, name := range AttorneyNames[:count] {
		lpa.Attorneys = append(lpa.Attorneys, MakeAttorney(name))
		firstNames = append(firstNames, name)
	}

	if count > 1 {
		lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
	}

	lpa.Tasks.ChooseAttorneys = TaskCompleted
	return firstNames
}

func AddReplacementAttorneys(lpa *Lpa, count int) []string {
	if count > len(ReplacementAttorneyNames) {
		count = len(ReplacementAttorneyNames)
	}

	var firstNames []string
	for _, name := range ReplacementAttorneyNames[:count] {
		lpa.ReplacementAttorneys = append(lpa.ReplacementAttorneys, MakeAttorney(name))
		firstNames = append(firstNames, name)
	}

	lpa.WantReplacementAttorneys = "yes"

	if count > 1 {
		lpa.HowReplacementAttorneysMakeDecisions = JointlyAndSeverally
		lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct
	}

	lpa.Tasks.ChooseReplacementAttorneys = TaskCompleted
	return firstNames
}

func CompleteHowAttorneysAct(lpa *Lpa, howTheyAct string) {
	switch howTheyAct {
	case Jointly:
		lpa.HowAttorneysMakeDecisions = Jointly
	case JointlyAndSeverally:
		lpa.HowAttorneysMakeDecisions = JointlyAndSeverally
	default:
		lpa.HowAttorneysMakeDecisions = JointlyForSomeSeverallyForOthers
		lpa.HowAttorneysMakeDecisionsDetails = "some details"
	}
}

func CompleteWhenCanLpaBeUsed(lpa *Lpa) {
	lpa.WhenCanTheLpaBeUsed = UsedWhenRegistered
	lpa.Tasks.WhenCanTheLpaBeUsed = TaskCompleted
}

func CompleteRestrictions(lpa *Lpa) {
	lpa.Restrictions = "Some restrictions on how Attorneys act"
	lpa.Tasks.Restrictions = TaskCompleted
}

func AddCertificateProvider(lpa *Lpa, firstNames string) {
	lpa.CertificateProvider = MakeCertificateProvider(firstNames)
	lpa.Tasks.CertificateProvider = TaskCompleted
}

func AddPeopleToNotify(lpa *Lpa, count int) []string {
	if count > len(PeopleToNotifyNames) {
		count = len(PeopleToNotifyNames)
	}

	var firstNames []string

	for _, name := range PeopleToNotifyNames[:count] {
		lpa.PeopleToNotify = append(lpa.PeopleToNotify, MakePersonToNotify(name))
		firstNames = append(firstNames, name)
	}

	lpa.DoYouWantToNotifyPeople = "yes"
	lpa.Tasks.PeopleToNotify = TaskCompleted

	return firstNames
}

func CompleteCheckYourLpa(lpa *Lpa) {
	lpa.Checked = true
	lpa.HappyToShare = true
	lpa.Tasks.CheckYourLpa = TaskCompleted
}

func PayForLpa(lpa *Lpa, store sesh.Store, r *http.Request, w http.ResponseWriter, ref string) {
	sesh.SetPayment(store, r, w, &sesh.PaymentSession{PaymentID: ref})

	lpa.PaymentDetails = PaymentDetails{
		PaymentReference: ref,
		PaymentId:        ref,
	}
	lpa.Tasks.PayForLpa = TaskCompleted
}

func ConfirmIdAndSign(lpa *Lpa) {
	lpa.OneLoginUserData = identity.UserData{
		OK:          true,
		RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
		FullName:    "Jose Smith",
	}

	lpa.WantToApplyForLpa = true
	lpa.WantToSignLpa = true
	lpa.Submitted = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
	lpa.CPWitnessCodeValidated = true
	lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted
}

func CompleteSectionOne(lpa *Lpa) {
	CompleteDonorDetails(lpa)
	AddAttorneys(lpa, 2)
	AddReplacementAttorneys(lpa, 2)
	CompleteWhenCanLpaBeUsed(lpa)
	CompleteRestrictions(lpa)
	AddCertificateProvider(lpa, "Barbara")
	AddPeopleToNotify(lpa, 2)
	CompleteCheckYourLpa(lpa)
}

func GetAttorneyByFirstNames(lpa *Lpa, firstNames string) (actor.Attorney, bool) {
	idx := slices.IndexFunc(lpa.Attorneys, func(a actor.Attorney) bool { return a.FirstNames == firstNames })
	if idx == -1 {
		return actor.Attorney{}, false
	}

	return lpa.Attorneys[idx], true
}

type fixtureData struct {
	App    AppData
	Errors validation.List
	Form   *fixturesForm
}

type fixturesForm struct {
	DonorDetails         string
	Attorneys            string
	ReplacementAttorneys string
	WhenCanLpaBeUsed     string
	Restrictions         string
	CertificateProvider  string
	PeopleToNotify       string
	PeopleToNotifyCount  string
	CheckAndSend         string
	Pay                  string
	IdAndSign            string
	CompleteAll          string
	Email                string
	CpFlowHasDonorPaid   string
}

func readFixtures(r *http.Request) *fixturesForm {
	return &fixturesForm{
		DonorDetails:         PostFormString(r, "donor-details"),
		Attorneys:            PostFormString(r, "choose-attorneys"),
		ReplacementAttorneys: PostFormString(r, "choose-replacement-attorneys"),
		WhenCanLpaBeUsed:     PostFormString(r, "when-can-lpa-be-used"),
		Restrictions:         PostFormString(r, "restrictions"),
		CertificateProvider:  PostFormString(r, "certificate-provider"),
		PeopleToNotify:       PostFormString(r, "people-to-notify"),
		PeopleToNotifyCount:  PostFormString(r, "ptn-count"),
		CheckAndSend:         PostFormString(r, "check-and-send-to-cp"),
		Pay:                  PostFormString(r, "pay-for-lpa"),
		IdAndSign:            PostFormString(r, "confirm-id-and-sign"),
		CompleteAll:          PostFormString(r, "complete-all-sections"),
		Email:                PostFormString(r, "email"),
		CpFlowHasDonorPaid:   PostFormString(r, "cp-flow-has-donor-paid"),
	}
}

func Fixtures(tmpl template.Template) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		data := &fixtureData{
			App:  appData,
			Form: &fixturesForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readFixtures(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				var values url.Values

				if data.Form.CpFlowHasDonorPaid != "" {
					values = url.Values{
						"useTestShareCode":           {"1"},
						data.Form.CpFlowHasDonorPaid: {"1"},
					}

					if data.Form.Email != "" {
						values.Add("withEmail", data.Form.Email)
					}
				} else {
					values = url.Values{
						data.Form.DonorDetails:         {"1"},
						data.Form.Attorneys:            {"1"},
						data.Form.ReplacementAttorneys: {"1"},
						data.Form.WhenCanLpaBeUsed:     {"1"},
						data.Form.Restrictions:         {"1"},
						data.Form.CertificateProvider:  {"1"},
						data.Form.CheckAndSend:         {"1"},
						data.Form.Pay:                  {"1"},
						data.Form.IdAndSign:            {"1"},
						data.Form.CompleteAll:          {"1"},
					}

					if data.Form.PeopleToNotify != "" {
						values.Add("withPeopleToNotify", data.Form.PeopleToNotifyCount)
					}
				}

				http.Redirect(w, r, fmt.Sprintf("%s?%s", Paths.TestingStart, values.Encode()), http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

func (f *fixturesForm) Validate() validation.List {
	var errors validation.List

	if f.Email != "" && f.CpFlowHasDonorPaid == "" {
		errors.String("cp-flow-has-donor-paid", "how to start the CP flow", f.CpFlowHasDonorPaid,
			validation.Select("startCpFlowWithId", "startCpFlowWithoutId"))
	}

	return errors
}
