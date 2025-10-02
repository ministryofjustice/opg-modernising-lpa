package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/names"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type WarningInterruptionData struct {
	App                 appcontext.Data
	Errors              validation.List
	Provided            *donordata.Provided
	Donor               *donordata.Donor
	Attorney            *donordata.Attorney
	ReplacementAttorney *donordata.Attorney
	CertificateProvider *donordata.CertificateProvider
	Correspondent       *donordata.Correspondent
	AuthorisedSignatory *donordata.AuthorisedSignatory
	IndependentWitness  *donordata.IndependentWitness
	PersonToNotify      *donordata.PersonToNotify
	Notifications       []page.Notification
	PageTitle           string
	From                string
	Next                string
}

func WarningInterruption(tmpl template.Template) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		actorType, err := actor.ParseType(r.FormValue("actor"))
		if err != nil {
			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		data := WarningInterruptionData{
			App:      appData,
			Provided: provided,
			From:     r.FormValue("warningFrom"),
			Next:     r.FormValue("next"),
		}

		switch actorType {
		case actor.TypeDonor:
			data.Donor = &provided.Donor
			data.PageTitle = "checkYourDetails"

			matches := donorMatches(provided, provided.Donor.FirstNames, provided.Donor.LastName)

			if provided.CertificateProvider.Address.Line1 != "" && provided.Donor.Address == provided.CertificateProvider.Address {
				matches = actor.TypeCertificateProvider
			}

			nameWarning := actor.NewSameNameWarning(
				actor.TypeDonor,
				matches,
				provided.Donor.FullName(),
			)
			if nameWarning != nil {
				data.Notifications = append(data.Notifications, page.Notification{
					Heading:  "pleaseReviewTheInformationYouHaveEntered",
					BodyHTML: nameWarning.Format(appData.Localizer),
				})
			}

			dobWarning := dateOfBirthWarning(provided.Donor.DateOfBirth, actor.TypeDonor)
			if dobWarning != "" {
				data.Notifications = append(data.Notifications, page.Notification{
					Heading:  "pleaseReviewTheInformationYouHaveEntered",
					BodyHTML: appData.Localizer.T(dobWarning),
				})
			}
		case actor.TypeAttorney, actor.TypeReplacementAttorney:
			uid, err := actoruid.Parse(r.FormValue("id"))
			if err != nil {
				return donor.PathTaskList.RedirectQuery(w, r, appData, provided, nil)
			}

			attorney, found := provided.Attorneys.Get(uid)
			matches := attorneyMatches(provided, uid, attorney.FirstNames, attorney.LastName)
			attorneyType := actor.TypeAttorney

			if actorType.IsReplacementAttorney() {
				attorney, found = provided.ReplacementAttorneys.Get(uid)
				matches = replacementAttorneyMatches(provided, uid, attorney.FirstNames, attorney.LastName)
				attorneyType = actor.TypeReplacementAttorney

				data.ReplacementAttorney = &attorney
				data.PageTitle = "checkYourReplacementAttorneysDetails"
			} else {
				data.Attorney = &attorney
				data.PageTitle = "checkYourAttorneysDetails"
			}

			if found {
				dobWarning := dateOfBirthWarning(attorney.DateOfBirth, attorneyType)
				if dobWarning != "" {
					data.Notifications = append(data.Notifications, page.Notification{
						Heading:  "pleaseReviewTheInformationYouHaveEntered",
						BodyHTML: appData.Localizer.Format(dobWarning, map[string]any{"FullName": attorney.FullName()}),
					})
				}

				nameWarning := actor.NewSameNameWarning(
					attorneyType,
					matches,
					attorney.FullName(),
				)
				if nameWarning != nil {
					data.Notifications = append(data.Notifications, page.Notification{
						Heading:  "pleaseReviewTheInformationYouHaveEntered",
						BodyHTML: nameWarning.Format(appData.Localizer),
					})

					// Warning can be triggered from name or address but we only give an option to change name in template
					from := donor.PathEnterAttorney
					if actorType.IsReplacementAttorney() {
						from = donor.PathEnterReplacementAttorney
					}

					data.From = from.FormatQuery(provided.LpaID, url.Values{"from": {data.Next}})
				}
			}
		case actor.TypeCertificateProvider:
			data.CertificateProvider = &provided.CertificateProvider
			data.PageTitle = "checkYourCertificateProvidersDetails"

			matches := certificateProviderMatches(provided, provided.CertificateProvider.FirstNames, provided.CertificateProvider.LastName)

			nameWarning := actor.NewSameNameWarning(
				actor.TypeCertificateProvider,
				matches,
				provided.CertificateProvider.FullName(),
			)
			if nameWarning != nil {
				data.Notifications = append(data.Notifications, page.Notification{
					Heading:  "pleaseReviewTheInformationYouHaveEntered",
					BodyHTML: nameWarning.Format(appData.Localizer),
				})
			}

			// pretend this page doesn't exist, then can use fromLink in the templates
			data.App.Page = data.Next
			data.From = ""

		case actor.TypeCorrespondent:
			data.Correspondent = &provided.Correspondent
			data.PageTitle = "checkYourCorrespondentsDetails"

			if correspondentNameMatchesDonor(provided, provided.Correspondent.FirstNames, provided.Correspondent.LastName) {
				data.Notifications = append(data.Notifications, page.Notification{
					Heading: "pleaseReviewTheInformationYouHaveEntered",
					BodyHTML: actor.NewSameNameWarning(
						actor.TypeCorrespondent,
						actor.TypeDonor,
						provided.Correspondent.FullName(),
					).Format(appData.Localizer),
				})
			}
		case actor.TypePersonToNotify:
			uid, err := actoruid.Parse(r.FormValue("id"))
			if err != nil {
				return donor.PathTaskList.RedirectQuery(w, r, appData, provided, nil)
			}

			person, found := provided.PeopleToNotify.Get(uid)
			if found {
				nameWarning := actor.NewSameNameWarning(
					actor.TypePersonToNotify,
					personToNotifyMatches(provided, uid, person.FirstNames, person.LastName),
					person.FullName(),
				)

				if nameWarning != nil {
					data.Notifications = append(data.Notifications, page.Notification{
						Heading:  "pleaseReviewTheInformationYouHaveEntered",
						BodyHTML: nameWarning.Format(appData.Localizer),
					})

					data.PageTitle = "checkYourPersonToNotifysDetails"
					data.PersonToNotify = &person

					from, _ := url.Parse(data.From)
					query := from.Query()
					query.Set("id", uid.String())
					from.RawQuery = query.Encode()

					data.From = from.String()
				}
			}
		case actor.TypeAuthorisedSignatory:
			data.AuthorisedSignatory = &provided.AuthorisedSignatory
			data.PageTitle = "checkYourAuthorisedSignatorysDetails"

			matches := signatoryMatches(provided, provided.AuthorisedSignatory.FirstNames, provided.AuthorisedSignatory.LastName)

			if !matches.IsNone() {
				data.Notifications = append(data.Notifications, page.Notification{
					Heading: "pleaseReviewTheInformationYouHaveEntered",
					BodyHTML: actor.NewSameNameWarning(
						actor.TypeAuthorisedSignatory,
						matches,
						provided.AuthorisedSignatory.FullName(),
					).Format(appData.Localizer),
				})
			}

		case actor.TypeIndependentWitness:
			data.IndependentWitness = &provided.IndependentWitness
			data.PageTitle = "checkYourIndependentWitnesssDetails"

			matches := independentWitnessMatches(provided, provided.IndependentWitness.FirstNames, provided.IndependentWitness.LastName)

			if !matches.IsNone() {
				data.Notifications = append(data.Notifications, page.Notification{
					Heading: "pleaseReviewTheInformationYouHaveEntered",
					BodyHTML: actor.NewSameNameWarning(
						actor.TypeIndependentWitness,
						matches,
						provided.IndependentWitness.FullName(),
					).Format(appData.Localizer),
				})
			}
		}

		if len(data.Notifications) == 0 {
			return donor.PathTaskList.RedirectQuery(w, r, appData, provided, nil)
		}

		return tmpl(w, data)
	}
}

func donorMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !person.Type.IsDonor() &&
			names.Equal(person.FirstNames, firstNames) &&
			names.Equal(person.LastName, lastName) {
			return person.Type
		}
	}

	if names.Equal(donor.Correspondent.FirstNames, firstNames) &&
		names.Equal(donor.Correspondent.LastName, lastName) {
		return actor.TypeCorrespondent
	}

	return actor.TypeNone
}

func attorneyMatches(donor *donordata.Provided, uid actoruid.UID, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !(person.Type.IsAttorney() && person.UID == uid) &&
			names.Equal(person.FirstNames, firstNames) &&
			names.Equal(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func replacementAttorneyMatches(donor *donordata.Provided, uid actoruid.UID, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !(person.Type.IsReplacementAttorney() && person.UID == uid) &&
			names.Equal(person.FirstNames, firstNames) &&
			names.Equal(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func certificateProviderMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if person.Type.IsCertificateProvider() || person.Type.IsPersonToNotify() {
			continue
		}

		if names.Equal(person.LastName, lastName) && names.Equal(person.FirstNames, firstNames) {
			return person.Type
		}

		if person.Type.IsAttorney() || person.Type.IsReplacementAttorney() || person.Type.IsDonor() {
			if names.Equal(person.LastName, lastName) && person.Address.Line1 != "" &&
				person.Address.Equal(donor.CertificateProvider.Address) {
				return person.Type
			}
		}
	}

	return actor.TypeNone
}

func correspondentNameMatchesDonor(donor *donordata.Provided, firstNames, lastName string) bool {
	return names.Equal(donor.Donor.FirstNames, firstNames) && names.Equal(donor.Donor.LastName, lastName)
}

func personToNotifyMatches(donor *donordata.Provided, uid actoruid.UID, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !(person.Type.IsPersonToNotify() && person.UID == uid) &&
			!person.Type.IsCertificateProvider() &&
			!person.Type.IsAuthorisedSignatory() &&
			!person.Type.IsIndependentWitness() &&
			names.Equal(person.FirstNames, firstNames) &&
			names.Equal(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func signatoryMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !person.Type.IsAuthorisedSignatory() &&
			!person.Type.IsPersonToNotify() &&
			names.Equal(person.FirstNames, firstNames) &&
			names.Equal(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func independentWitnessMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !person.Type.IsIndependentWitness() &&
			!person.Type.IsPersonToNotify() &&
			names.Equal(person.FirstNames, firstNames) &&
			names.Equal(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}

func dateOfBirthWarning(dateOfBirth date.Date, actorType actor.Type) string {
	var (
		today                = date.Today()
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !dateOfBirth.IsZero() {
		if dateOfBirth.Before(hundredYearsEarlier) {
			if actorType.IsAttorney() || actorType.IsReplacementAttorney() {
				return "dateOfBirthIsOver100AttorneyWarning"
			} else if actorType.IsDonor() {
				return "dateOfBirthIsOver100DonorWarning"
			}
		}

		if dateOfBirth.Before(today) && dateOfBirth.After(eighteenYearsEarlier) {
			if actorType.IsAttorney() || actorType.IsReplacementAttorney() {
				return "dateOfBirthIsUnder18AttorneyWarning"
			}
		}
	}

	return ""
}
