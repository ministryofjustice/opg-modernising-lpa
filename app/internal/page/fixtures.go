package page

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"golang.org/x/exp/slices"
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
	donor := actor.Donor{
		FirstNames: "Jamie",
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

	return donor
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
		LastName:                "Jones",
		Email:                   TestEmail,
		Mobile:                  TestMobile,
		DateOfBirth:             date.New("1997", "1", "2"),
		Relationship:            "friend",
		RelationshipDescription: "",
		RelationshipLength:      "gte-2-years",
		CarryOutBy:              "paper",
		Address: place.Address{
			Line1:      "5 RICHMOND PLACE",
			Line2:      "KINGS HEATH",
			Line3:      "WEST MIDLANDS",
			TownOrCity: "BIRMINGHAM",
			Postcode:   "B14 7ED",
		},
	}
}

func CompleteDonorDetails(lpa *Lpa) {
	lpa.Donor = MakePerson()
	lpa.WhoFor = "me"
	lpa.Type = "pfa"
	lpa.Tasks.YourDetails = actor.TaskCompleted
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
		lpa.AttorneyDecisions.How = actor.JointlyAndSeverally
	}

	lpa.Tasks.ChooseAttorneys = actor.TaskCompleted
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
		lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally
		lpa.HowShouldReplacementAttorneysStepIn = OneCanNoLongerAct
	}

	lpa.Tasks.ChooseReplacementAttorneys = actor.TaskCompleted
	return firstNames
}

func CompleteHowAttorneysAct(lpa *Lpa, howTheyAct string) {
	switch howTheyAct {
	case actor.Jointly:
		lpa.AttorneyDecisions.How = actor.Jointly
	case actor.JointlyAndSeverally:
		lpa.AttorneyDecisions.How = actor.JointlyAndSeverally
	default:
		lpa.AttorneyDecisions.How = actor.JointlyForSomeSeverallyForOthers
		lpa.AttorneyDecisions.Details = "some details"
	}
}

func CompleteHowReplacementAttorneysAct(lpa *Lpa, howTheyAct string) {
	switch howTheyAct {
	case actor.Jointly:
		lpa.ReplacementAttorneyDecisions.How = actor.Jointly
		lpa.ReplacementAttorneyDecisions.HappyIfOneCannotActNoneCan = "yes"
	case actor.JointlyAndSeverally:
		lpa.ReplacementAttorneyDecisions.How = actor.JointlyAndSeverally
		lpa.ReplacementAttorneyDecisions.HappyIfOneCannotActNoneCan = "yes"
	default:
		lpa.ReplacementAttorneyDecisions.How = actor.JointlyForSomeSeverallyForOthers
		lpa.ReplacementAttorneyDecisions.HappyIfOneCannotActNoneCan = "yes"
		lpa.ReplacementAttorneyDecisions.Details = "some details"
	}
}

func CompleteWhenCanLpaBeUsed(lpa *Lpa) {
	lpa.WhenCanTheLpaBeUsed = UsedWhenRegistered
	lpa.Tasks.WhenCanTheLpaBeUsed = actor.TaskCompleted
}

func CompleteRestrictions(lpa *Lpa) {
	lpa.Restrictions = "My attorneys must not sell my home unless, in my doctor’s opinion, I can no longer live independently"
	lpa.Tasks.Restrictions = actor.TaskCompleted
}

func AddCertificateProvider(lpa *Lpa, firstNames string) {
	lpa.CertificateProvider = MakeCertificateProvider(firstNames)
	lpa.Tasks.CertificateProvider = actor.TaskCompleted
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
	lpa.Tasks.PeopleToNotify = actor.TaskCompleted

	return firstNames
}

func CompleteCheckYourLpa(lpa *Lpa) {
	lpa.Checked = true
	lpa.HappyToShare = true
	lpa.Tasks.CheckYourLpa = actor.TaskCompleted
}

func PayForLpa(lpa *Lpa, store sesh.Store, r *http.Request, w http.ResponseWriter, ref string) {
	sesh.SetPayment(store, r, w, &sesh.PaymentSession{PaymentID: ref})

	lpa.PaymentDetails = PaymentDetails{
		PaymentReference: ref,
		PaymentId:        ref,
	}
	lpa.Tasks.PayForLpa = actor.TaskCompleted
}

func ConfirmIdAndSign(lpa *Lpa) {
	lpa.DonorIdentityUserData = identity.UserData{
		OK:          true,
		Provider:    identity.OneLogin,
		RetrievedAt: time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC),
		FirstNames:  "Jamie",
		LastName:    "Smith",
	}

	lpa.WantToApplyForLpa = true
	lpa.WantToSignLpa = true
	lpa.Submitted = time.Date(2023, time.January, 2, 3, 4, 5, 6, time.UTC)
	lpa.CPWitnessCodeValidated = true
	lpa.Tasks.ConfirmYourIdentityAndSign = actor.TaskCompleted
}

func CompleteSectionOne(lpa *Lpa) {
	CompleteDonorDetails(lpa)
	AddAttorneys(lpa, 2)
	AddReplacementAttorneys(lpa, 2)
	CompleteWhenCanLpaBeUsed(lpa)
	CompleteRestrictions(lpa)
	AddCertificateProvider(lpa, "Jessie")
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
	Journey                string
	DonorDetails           string
	Attorneys              string
	ReplacementAttorneys   string
	WhenCanLpaBeUsed       string
	Restrictions           string
	CertificateProvider    string
	PeopleToNotify         string
	PeopleToNotifyCount    string
	CheckAndSend           string
	Pay                    string
	IdAndSign              string
	CompleteAll            string
	Email                  string
	CpFlowHasDonorPaid     string
	ForReplacementAttorney string
	Signed                 string
	Type                   string
}

func readFixtures(r *http.Request) *fixturesForm {
	return &fixturesForm{
		Journey:                PostFormString(r, "journey"),
		DonorDetails:           PostFormString(r, "donor-details"),
		Attorneys:              PostFormString(r, "choose-attorneys"),
		ReplacementAttorneys:   PostFormString(r, "choose-replacement-attorneys"),
		WhenCanLpaBeUsed:       PostFormString(r, "when-can-lpa-be-used"),
		Restrictions:           PostFormString(r, "restrictions"),
		CertificateProvider:    PostFormString(r, "certificate-provider"),
		PeopleToNotify:         PostFormString(r, "people-to-notify"),
		PeopleToNotifyCount:    PostFormString(r, "ptn-count"),
		CheckAndSend:           PostFormString(r, "check-and-send-to-cp"),
		Pay:                    PostFormString(r, "pay-for-lpa"),
		IdAndSign:              PostFormString(r, "confirm-id-and-sign"),
		CompleteAll:            PostFormString(r, "complete-all-sections"),
		Email:                  PostFormString(r, "email"),
		CpFlowHasDonorPaid:     PostFormString(r, "cp-flow-has-donor-paid"),
		ForReplacementAttorney: PostFormString(r, "for-replacement-attorney"),
		Signed:                 PostFormString(r, "signed"),
		Type:                   PostFormString(r, "type"),
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

				switch data.Form.Journey {
				case "attorney":
					values = url.Values{
						"useTestShareCode":           {"1"},
						"sendAttorneyShare":          {"1"},
						"completeLpa":                {"1"},
						"withAttorneys":              {"1"},
						"howAttorneysAct":            {"jointly-and-severally"},
						"withReplacementAttorneys":   {"1"},
						"howReplacementAttorneysAct": {"jointly"},
						"withType":                   {data.Form.Type},
						"withRestrictions":           {"1"},
						"redirect":                   {Paths.Attorney.Start},
					}
					if data.Form.Email != "" {
						values.Add("withEmail", data.Form.Email)
					}
					if data.Form.ForReplacementAttorney != "" {
						values.Add("forReplacementAttorney", "1")
					}
					if data.Form.Signed != "" {
						values.Add("signedByDonor", "1")
						values.Add("provideCertificate", "1")
					}

				case "certificate-provider":
					values = url.Values{
						"useTestShareCode":           {"1"},
						data.Form.CpFlowHasDonorPaid: {"1"},
					}

					if data.Form.Email != "" {
						values.Add("withEmail", data.Form.Email)
					}
				case "donor":
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

	if f.Journey == "certificate-provider" && f.Email != "" && f.CpFlowHasDonorPaid == "" {
		errors.String("cp-flow-has-donor-paid", "how to start the CP flow", f.CpFlowHasDonorPaid,
			validation.Select("startCpFlowWithId", "startCpFlowWithoutId"))
	}

	return errors
}
