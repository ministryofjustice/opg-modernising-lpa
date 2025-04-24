package donorpage

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/stretchr/testify/assert"
)

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

//func TestPostChooseAttorneysWhenNameWarning(t *testing.T) {
//	form := url.Values{
//		"first-names":         {"Jane"},
//		"last-name":           {"Doe"},
//		"email":               {"name@example.com"},
//		"date-of-birth-day":   {"2"},
//		"date-of-birth-month": {"1"},
//		"date-of-birth-year":  {"1990"},
//	}
//
//	w := httptest.NewRecorder()
//	r, _ := http.NewRequest(http.MethodPost, "/?id="+testUID.String(), strings.NewReader(form.Encode()))
//	r.Header.Add("Content-Type", page.FormUrlEncoded)
//
//	donorStore := newMockDonorStore(t)
//	donorStore.EXPECT().
//		Put(r.Context(), mock.Anything).
//		Return(nil)
//
//	localizer := newMockLocalizer(t)
//	localizer.EXPECT().
//		T("attorney").
//		Return("1").
//		Once()
//	localizer.EXPECT().
//		T("attorneys").
//		Return("2").
//		Once()
//	localizer.EXPECT().
//		T("anAttorney").
//		Return("3").
//		Once()
//	localizer.EXPECT().
//		Format(
//			"donorMatchesActorWarning",
//			map[string]any{"ArticleAndType": "3", "FirstNames": "Jane", "LastName": "Doe", "Type": "1", "TypePlural": "2"}).
//		Return("translatedWarning").
//		Once()
//
//	testAppData.Page = "/a"
//	testAppData.Localizer = localizer
//
//	err := ChooseAttorneys(nil, donorStore)(testAppData, w, r, &donordata.Provided{
//		LpaID: "lpa-id",
//		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
//	})
//	resp := w.Result()
//
//	assert.Nil(t, err)
//	assert.Equal(t, donor.PathWarningInterruption.Format("lpa-id")+"?id="+testUID.String()+"&warningFrom=%2Fa", resp.Header.Get("Location"))
//}
