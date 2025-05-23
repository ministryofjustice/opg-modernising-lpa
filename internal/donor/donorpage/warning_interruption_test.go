package donorpage

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetWarningInterruptionChooseAttorneys(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathEnterAttorney.Format("lpa-id")},
		"id":          {testUID.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("theDonor").
		Return("1").
		Once()
	localizer.EXPECT().
		T("attorney").
		Return("2").
		Once()
	localizer.EXPECT().
		T("attorneys").
		Return("3").
		Once()
	localizer.EXPECT().
		T("anAttorney").
		Return("4").
		Once()
	localizer.EXPECT().
		Format(
			"donorMatchesActorWarning",
			map[string]any{"ArticleAndType": "4", "FirstNames": "Jane", "LastName": "Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3"}).
		Return("translatedWarning").
		Once()

	yesterday := date.Today().AddDate(0, 0, -1)
	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		Attorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{
				{UID: testUID, FirstNames: "Jane", LastName: "Doe", DateOfBirth: yesterday},
			},
		},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:      appData,
			Donor:    provided,
			Attorney: donordata.Attorney{UID: testUID, FirstNames: "Jane", LastName: "Doe", DateOfBirth: yesterday},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "dateOfBirthIsUnder18Attorney"},
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourAttorneysDetails",
			From:      donor.PathEnterAttorney.Format("lpa-id"),
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionWhenTemplateError(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathEnterAttorney.Format("lpa-id")},
		"id":          {testUID.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("1").
		Times(4)
	localizer.EXPECT().
		Format(mock.Anything, mock.Anything).
		Return("translatedWarning").
		Once()

	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		Attorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{
				{UID: testUID, FirstNames: "Jane", LastName: "Doe"},
			},
		},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionWhenCantShowWarnings(t *testing.T) {
	uid := actoruid.New()

	testcases := map[string]struct {
		queryUID, warningFrom string
	}{
		"path not supported": {
			warningFrom: donor.PathTaskList.Format("lpa-id"),
		},
		"not a UID": {
			queryUID:    "not a UID",
			warningFrom: donor.PathEnterAttorney.Format("lpa-id"),
		},
		"attorney not found": {
			queryUID:    uid.String(),
			warningFrom: donor.PathEnterAttorney.Format("lpa-id"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			query := url.Values{
				"warningFrom": {tc.warningFrom},
				"id":          {tc.queryUID},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

			err := WarningInterruption(nil)(testAppData, w, r, &donordata.Provided{
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{
						{UID: testUID, FirstNames: "Jane", LastName: "Doe"},
					},
				}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
		})
	}
}

func TestDateOfBirthWarning(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		dob           date.Date
		warning       string
		isReplacement bool
	}{
		"valid": {
			dob: validDob,
		},
		"future dob": {
			dob: now.AddDate(0, 0, 1),
		},
		"dob is 18": {
			dob: now.AddDate(-18, 0, 0),
		},
		"dob under 18": {
			dob:     now.AddDate(-18, 0, 2),
			warning: "dateOfBirthIsUnder18Attorney",
		},
		"dob under 18 replacement": {
			dob:           now.AddDate(-18, 0, 2),
			warning:       "attorneyDateOfBirthIsUnder18",
			isReplacement: true,
		},
		"dob is 100": {
			dob: now.AddDate(-100, 0, 0),
		},
		"dob over 100": {
			dob:     now.AddDate(-100, 0, -1),
			warning: "dateOfBirthIsOver100",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.warning, DateOfBirthWarning(tc.dob, tc.isReplacement))
		})
	}
}

func TestAttorneyMatches(t *testing.T) {
	uid := actoruid.New()

	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{UID: uid, FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  donordata.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, uid, "x", "y"))
	assert.Equal(t, actor.TypeDonor, attorneyMatches(donor, uid, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, attorneyMatches(donor, uid, "c", "d"))
	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, uid, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, attorneyMatches(donor, uid, "g", "h"))
	assert.Equal(t, actor.TypeReplacementAttorney, attorneyMatches(donor, uid, "I", "J"))
	assert.Equal(t, actor.TypeCertificateProvider, attorneyMatches(donor, uid, "k", "l"))
	assert.Equal(t, actor.TypePersonToNotify, attorneyMatches(donor, uid, "M", "N"))
	assert.Equal(t, actor.TypePersonToNotify, attorneyMatches(donor, uid, "o", "p"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, attorneyMatches(donor, uid, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, attorneyMatches(donor, uid, "i", "w"))
}

func TestAttorneyMatchesEmptyNamesIgnored(t *testing.T) {
	uid := actoruid.New()

	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{UID: uid, FirstNames: "", LastName: ""},
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, attorneyMatches(donor, uid, "", ""))
}
