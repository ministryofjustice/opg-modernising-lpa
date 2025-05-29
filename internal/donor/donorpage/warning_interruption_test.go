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

func TestGetWarningInterruptionDonor(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathYourDetails.Format("lpa-id")},
		"next":        {"/next-page"},
		"actor":       {actor.TypeDonor.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("anAttorney").
		Return("1").
		Once()
	localizer.EXPECT().
		T("donor").
		Return("2").
		Once()
	localizer.EXPECT().
		T("").
		Return("3").
		Once()
	localizer.EXPECT().
		T("theDonor").
		Return("4").
		Once()
	localizer.EXPECT().
		T("attorney").
		Return("5").
		Once()
	localizer.EXPECT().
		T("dateOfBirthIsOver100DonorWarning").
		Return("translatedDobWarning").
		Once()
	localizer.EXPECT().
		Format(
			"donorMatchesActorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
		Return("translatedWarning").
		Once()

	oneHundredOneYearsAgo := date.Today().AddDate(-101, 0, 0)
	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe", DateOfBirth: oneHundredOneYearsAgo},
		Attorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{
				{UID: testUID, FirstNames: "Jane", LastName: "Doe"},
			},
		},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:      appData,
			Provided: provided,
			Donor:    &donordata.Donor{FirstNames: "Jane", LastName: "Doe", DateOfBirth: oneHundredOneYearsAgo},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedDobWarning"},
			},
			PageTitle: "checkYourDetails",
			From:      donor.PathYourDetails.Format("lpa-id"),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionAttorney(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathEnterAttorney.Format("lpa-id")},
		"id":          {testUID.String()},
		"next":        {"/next-page"},
		"actor":       {actor.TypeAttorney.String()},
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
		T("donor").
		Return("5").
		Once()
	localizer.EXPECT().
		Format(
			"dateOfBirthIsUnder18AttorneyWarning",
			map[string]any{"FullName": "Jane Doe"}).
		Return("translatedDobWarning").
		Once()
	localizer.EXPECT().
		Format(
			"actorMatchesDonorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
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
			Provided: provided,
			Attorney: &donordata.Attorney{UID: testUID, FirstNames: "Jane", LastName: "Doe", DateOfBirth: yesterday},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedDobWarning"},
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourAttorneysDetails",
			From:      donor.PathEnterAttorney.Format("lpa-id"),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionReplacementAttorney(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathChooseReplacementAttorneys.Format("lpa-id")},
		"id":          {testUID.String()},
		"next":        {"/next-page"},
		"actor":       {actor.TypeReplacementAttorney.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("theDonor").
		Return("1").
		Once()
	localizer.EXPECT().
		T("replacementAttorney").
		Return("2").
		Once()
	localizer.EXPECT().
		T("replacementAttorneys").
		Return("3").
		Once()
	localizer.EXPECT().
		T("aReplacementAttorney").
		Return("4").
		Once()
	localizer.EXPECT().
		T("donor").
		Return("5").
		Once()
	localizer.EXPECT().
		Format(
			"dateOfBirthIsUnder18AttorneyWarning",
			map[string]any{"FullName": "Jane Doe"}).
		Return("translatedDobWarning").
		Once()
	localizer.EXPECT().
		Format(
			"actorMatchesDonorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
		Return("translatedWarning").
		Once()

	yesterday := date.Today().AddDate(0, 0, -1)
	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		ReplacementAttorneys: donordata.Attorneys{
			Attorneys: []donordata.Attorney{
				{UID: testUID, FirstNames: "Jane", LastName: "Doe", DateOfBirth: yesterday},
			},
		},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:                 appData,
			Provided:            provided,
			ReplacementAttorney: &donordata.Attorney{UID: testUID, FirstNames: "Jane", LastName: "Doe", DateOfBirth: yesterday},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedDobWarning"},
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourReplacementAttorneysDetails",
			From:      donor.PathChooseReplacementAttorneys.Format("lpa-id"),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionCertificateProvider(t *testing.T) {
	testcases := map[string]*donordata.Provided{
		"name matches": {
			LpaID:               "lpa-id",
			Donor:               donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
			CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Doe"},
		},
		"address matches": {
			LpaID:               "lpa-id",
			Donor:               donordata.Donor{FirstNames: "John", LastName: "Doe", Address: testAddress},
			CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Doe", Address: testAddress},
		},
	}

	for name, donorProvided := range testcases {
		t.Run(name, func(t *testing.T) {
			query := url.Values{
				"warningFrom": {donor.PathCertificateProviderDetails.Format("lpa-id")},
				"next":        {"/next-page"},
				"actor":       {actor.TypeCertificateProvider.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("theDonor").
				Return("1").
				Once()
			localizer.EXPECT().
				T("certificateProvider").
				Return("2").
				Once()
			localizer.EXPECT().
				T("certificateProviders").
				Return("3").
				Once()
			localizer.EXPECT().
				T("theCertificateProvider").
				Return("4").
				Once()
			localizer.EXPECT().
				T("donor").
				Return("5").
				Once()
			localizer.EXPECT().
				Format(
					"actorMatchesDonorNameOrAddressWarning",
					map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
				Return("translatedWarning").
				Once()

			appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, WarningInterruptionData{
					App:                 appData,
					Provided:            donorProvided,
					CertificateProvider: &donorProvided.CertificateProvider,
					Notifications: []page.Notification{
						{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
					},
					PageTitle: "checkYourCertificateProvidersDetails",
					From:      donor.PathCertificateProviderDetails.Format("lpa-id"),
					Next:      "/next-page",
				}).
				Return(nil)

			err := WarningInterruption(template.Execute)(appData, w, r, donorProvided)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetWarningInterruptionEnterCorrespondentDetails(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathEnterCorrespondentDetails.Format("lpa-id")},
		"next":        {"/next-page"},
		"actor":       {actor.TypeCorrespondent.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("theDonor").
		Return("1").
		Once()
	localizer.EXPECT().
		T("correspondent").
		Return("2").
		Once()
	localizer.EXPECT().
		T("").
		Return("3").
		Once()
	localizer.EXPECT().
		T("theCorrespondent").
		Return("4").
		Once()
	localizer.EXPECT().
		T("donor").
		Return("5").
		Once()
	localizer.EXPECT().
		Format(
			"correspondentMatchesDonorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
		Return("translatedWarning").
		Once()

	provided := &donordata.Provided{
		LpaID:         "lpa-id",
		Donor:         donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		Correspondent: donordata.Correspondent{FirstNames: "Jane", LastName: "Doe"},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:           appData,
			Provided:      provided,
			Correspondent: &donordata.Correspondent{FirstNames: "Jane", LastName: "Doe"},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourCorrespondentsDetails",
			From:      donor.PathEnterCorrespondentDetails.Format("lpa-id"),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionChoosePeopleToNotify(t *testing.T) {
	query := url.Values{
		"id":          {testUID.String()},
		"warningFrom": {donor.PathChoosePeopleToNotify.Format("lpa-id")},
		"next":        {"/next-page"},
		"actor":       {actor.TypePersonToNotify.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("theDonor").
		Return("1").
		Once()
	localizer.EXPECT().
		T("personToNotify").
		Return("2").
		Once()
	localizer.EXPECT().
		T("peopleToNotify").
		Return("3").
		Once()
	localizer.EXPECT().
		T("aPersonToNotify").
		Return("4").
		Once()
	localizer.EXPECT().
		T("donor").
		Return("5").
		Once()
	localizer.EXPECT().
		Format(
			"personToNotifyMatchesDonorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
		Return("translatedWarning").
		Once()

	provided := &donordata.Provided{
		LpaID: "lpa-id",
		Donor: donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		PeopleToNotify: donordata.PeopleToNotify{
			donordata.PersonToNotify{UID: testUID, FirstNames: "Jane", LastName: "Doe"},
		},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:            appData,
			Provided:       provided,
			PersonToNotify: &donordata.PersonToNotify{UID: testUID, FirstNames: "Jane", LastName: "Doe"},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourPersonToNotifysDetails",
			From:      donor.PathChoosePeopleToNotify.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionYourAuthorisedSignatory(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathYourAuthorisedSignatory.Format("lpa-id")},
		"next":        {"/next-page"},
		"actor":       {actor.TypeAuthorisedSignatory.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("theDonor").
		Return("1").
		Once()
	localizer.EXPECT().
		T("signatory").
		Return("2").
		Once()
	localizer.EXPECT().
		T("").
		Return("3").
		Once()
	localizer.EXPECT().
		T("theAuthorisedSignatory").
		Return("4").
		Once()
	localizer.EXPECT().
		T("donor").
		Return("5").
		Once()
	localizer.EXPECT().
		Format(
			"actorMatchesDonorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
		Return("translatedWarning").
		Once()

	provided := &donordata.Provided{
		LpaID:               "lpa-id",
		Donor:               donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "Jane", LastName: "Doe"},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:                 appData,
			Provided:            provided,
			AuthorisedSignatory: &donordata.AuthorisedSignatory{FirstNames: "Jane", LastName: "Doe"},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourAuthorisedSignatorysDetails",
			From:      donor.PathYourAuthorisedSignatory.Format("lpa-id"),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionIndependentWitness(t *testing.T) {
	query := url.Values{
		"warningFrom": {donor.PathYourIndependentWitness.Format("lpa-id")},
		"next":        {"/next-page"},
		"actor":       {actor.TypeIndependentWitness.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("theDonor").
		Return("1").
		Once()
	localizer.EXPECT().
		T("independentWitness").
		Return("2").
		Once()
	localizer.EXPECT().
		T("").
		Return("3").
		Once()
	localizer.EXPECT().
		T("theIndependentWitness").
		Return("4").
		Once()
	localizer.EXPECT().
		T("donor").
		Return("5").
		Once()
	localizer.EXPECT().
		Format(
			"actorMatchesDonorNameWarning",
			map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
		Return("translatedWarning").
		Once()

	provided := &donordata.Provided{
		LpaID:              "lpa-id",
		Donor:              donordata.Donor{FirstNames: "Jane", LastName: "Doe"},
		IndependentWitness: donordata.IndependentWitness{FirstNames: "Jane", LastName: "Doe"},
	}

	appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, WarningInterruptionData{
			App:                appData,
			Provided:           provided,
			IndependentWitness: &donordata.IndependentWitness{FirstNames: "Jane", LastName: "Doe"},
			Notifications: []page.Notification{
				{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
			},
			PageTitle: "checkYourIndependentWitnesssDetails",
			From:      donor.PathYourIndependentWitness.Format("lpa-id"),
			Next:      "/next-page",
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
		"actor":       {actor.TypeAttorney.String()},
	}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("1").
		Times(5)
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
		"no name match": {
			warningFrom: donor.PathCertificateProviderDetails.Format("lpa-id"),
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
				LpaID: "lpa-id",
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{
						{UID: testUID, FirstNames: "Jane", LastName: "Doe"},
					},
				}})
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, donor.PathTaskList.Format("lpa-id"), resp.Header.Get("Location"))
		})
	}
}

func TestDateOfBirthWarning(t *testing.T) {
	now := date.Today()
	validDob := now.AddDate(-18, 0, -1)

	testCases := map[string]struct {
		dob       date.Date
		warning   string
		actorType actor.Type
	}{
		"valid": {
			dob:       validDob,
			actorType: actor.TypeAttorney,
		},
		"future dob": {
			dob:       now.AddDate(0, 0, 1),
			actorType: actor.TypeAttorney,
		},
		"dob is 18": {
			dob:       now.AddDate(-18, 0, 0),
			actorType: actor.TypeAttorney,
		},
		"dob under 18": {
			dob:       now.AddDate(-18, 0, 2),
			warning:   "dateOfBirthIsUnder18AttorneyWarning",
			actorType: actor.TypeAttorney,
		},
		"dob under 18 replacement": {
			dob:       now.AddDate(-18, 0, 2),
			warning:   "dateOfBirthIsUnder18AttorneyWarning",
			actorType: actor.TypeReplacementAttorney,
		},
		"dob is 100": {
			dob:       now.AddDate(-100, 0, 0),
			actorType: actor.TypeAttorney,
		},
		"dob over 100": {
			dob:       now.AddDate(-100, 0, -1),
			warning:   "dateOfBirthIsOver100AttorneyWarning",
			actorType: actor.TypeAttorney,
		},
		"dob over 100 replacement": {
			dob:       now.AddDate(-100, 0, -1),
			warning:   "dateOfBirthIsOver100AttorneyWarning",
			actorType: actor.TypeReplacementAttorney,
		},
		"dob over 100 donor": {
			dob:       now.AddDate(-100, 0, -1),
			warning:   "dateOfBirthIsOver100DonorWarning",
			actorType: actor.TypeDonor,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.warning, dateOfBirthWarning(tc.dob, tc.actorType))
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

func TestCertificateProviderMatches(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
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

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeDonor, certificateProviderMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(donor, "c", "d"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(donor, "E", "F"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(donor, "g", "h"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(donor, "I", "J"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "o", "p"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, certificateProviderMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, certificateProviderMatches(donor, "i", "w"))
}

func TestCertificateProviderMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
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

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "", ""))
}

func TestReplacementAttorneyMatches(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{UID: uid, FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{FirstNames: "o", LastName: "p"},
		},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  donordata.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(donor, uid, "x", "y"))
	assert.Equal(t, actor.TypeDonor, replacementAttorneyMatches(donor, uid, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, replacementAttorneyMatches(donor, uid, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, replacementAttorneyMatches(donor, uid, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, replacementAttorneyMatches(donor, uid, "g", "h"))
	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(donor, uid, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, replacementAttorneyMatches(donor, uid, "K", "l"))
	assert.Equal(t, actor.TypePersonToNotify, replacementAttorneyMatches(donor, uid, "m", "n"))
	assert.Equal(t, actor.TypePersonToNotify, replacementAttorneyMatches(donor, uid, "O", "P"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, replacementAttorneyMatches(donor, uid, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, replacementAttorneyMatches(donor, uid, "i", "w"))
}

func TestReplacementAttorneyMatchesEmptyNamesIgnored(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
			{UID: uid, FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, replacementAttorneyMatches(donor, uid, "", ""))
}

func TestPersonToNotifyMatches(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "g", LastName: "h"},
			{FirstNames: "i", LastName: "j"},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "k", LastName: "l"},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "m", LastName: "n"},
			{UID: uid, FirstNames: "o", LastName: "p"},
		},
	}

	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "x", "y"))
	assert.Equal(t, actor.TypeDonor, personToNotifyMatches(donor, uid, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, personToNotifyMatches(donor, uid, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, personToNotifyMatches(donor, uid, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, personToNotifyMatches(donor, uid, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, personToNotifyMatches(donor, uid, "i", "j"))
	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "k", "L"))
	assert.Equal(t, actor.TypePersonToNotify, personToNotifyMatches(donor, uid, "m", "n"))
	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "o", "p"))
}

func TestPersonToNotifyMatchesEmptyNamesIgnored(t *testing.T) {
	uid := actoruid.New()
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "", LastName: ""},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "", LastName: ""},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "", LastName: ""},
			{UID: uid, FirstNames: "", LastName: ""},
		},
	}

	assert.Equal(t, actor.TypeNone, personToNotifyMatches(donor, uid, "", ""))
}

func TestSignatoryMatches(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
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

	assert.Equal(t, actor.TypeNone, signatoryMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeDonor, signatoryMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, signatoryMatches(donor, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, signatoryMatches(donor, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, signatoryMatches(donor, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, signatoryMatches(donor, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, signatoryMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypeNone, signatoryMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypeNone, signatoryMatches(donor, "O", "P"))
	assert.Equal(t, actor.TypeNone, signatoryMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, signatoryMatches(donor, "i", "w"))
}

func TestSignatoryMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &donordata.Provided{
		Attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		PeopleToNotify:       donordata.PeopleToNotify{{}},
	}

	assert.Equal(t, actor.TypeNone, signatoryMatches(donor, "", ""))
}

func TestIndependentWitnessMatches(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
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

	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeDonor, independentWitnessMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, independentWitnessMatches(donor, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, independentWitnessMatches(donor, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, independentWitnessMatches(donor, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, independentWitnessMatches(donor, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, independentWitnessMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "O", "P"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, independentWitnessMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "i", "w"))
}

func TestIndependentWitnessMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &donordata.Provided{
		Attorneys:            donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{{}}},
		PeopleToNotify:       donordata.PeopleToNotify{{}},
	}

	assert.Equal(t, actor.TypeNone, independentWitnessMatches(donor, "", ""))
}

func TestDonorMatches(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b"},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d"},
			{FirstNames: "e", LastName: "f"},
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

	assert.Equal(t, actor.TypeNone, donorMatches(donor, "x", "y"))
	assert.Equal(t, actor.TypeNone, donorMatches(donor, "a", "b"))
	assert.Equal(t, actor.TypeAttorney, donorMatches(donor, "C", "D"))
	assert.Equal(t, actor.TypeAttorney, donorMatches(donor, "e", "f"))
	assert.Equal(t, actor.TypeReplacementAttorney, donorMatches(donor, "G", "H"))
	assert.Equal(t, actor.TypeReplacementAttorney, donorMatches(donor, "i", "j"))
	assert.Equal(t, actor.TypeCertificateProvider, donorMatches(donor, "k", "l"))
	assert.Equal(t, actor.TypePersonToNotify, donorMatches(donor, "m", "n"))
	assert.Equal(t, actor.TypePersonToNotify, donorMatches(donor, "O", "P"))
	assert.Equal(t, actor.TypeAuthorisedSignatory, donorMatches(donor, "a", "s"))
	assert.Equal(t, actor.TypeIndependentWitness, donorMatches(donor, "i", "w"))
}

func TestDonorMatchesEmptyNamesIgnored(t *testing.T) {
	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "", LastName: ""},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
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

	assert.Equal(t, actor.TypeNone, donorMatches(donor, "", ""))
}
