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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
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
		"warningFrom": {donor.PathChooseAttorneysAddress.Format("lpa-id")},
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
			From:      donor.PathEnterAttorney.FormatQuery("lpa-id", url.Values{"from": {"/next-page"}}),
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
		"warningFrom": {donor.PathChooseReplacementAttorneysAddress.Format("lpa-id")},
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
			From:      donor.PathEnterReplacementAttorney.FormatQuery("lpa-id", url.Values{"from": {"/next-page"}}),
			Next:      "/next-page",
		}).
		Return(nil)

	err := WarningInterruption(template.Execute)(appData, w, r, provided)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWarningInterruptionCertificateProvider(t *testing.T) {
	testcases := map[string]struct {
		donorProvided         *donordata.Provided
		match                 string
		matchArticleAndType   string
		warningTranslationKey string
	}{
		"donor": {
			donorProvided: &donordata.Provided{
				LpaID:               "lpa-id",
				Donor:               donordata.Donor{FirstNames: "John", LastName: "Doe", Address: testAddress},
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Doe", Address: testAddress},
			},
			match:                 "donor",
			matchArticleAndType:   "theDonor",
			warningTranslationKey: "actorMatchesDonorNameOrAddressWarning",
		},
		"attorney": {
			donorProvided: &donordata.Provided{
				LpaID:               "lpa-id",
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Doe", Address: testAddress},
				Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
					{FirstNames: "John", LastName: "Doe", Address: testAddress},
				}},
			},
			match:                 "attorney",
			matchArticleAndType:   "anAttorney",
			warningTranslationKey: "actorMatchesDifferentActorNameOrAddressWarningConfirmLater",
		},
		"replacement attorney": {
			donorProvided: &donordata.Provided{
				LpaID:               "lpa-id",
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Doe", Address: testAddress},
				ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
					{FirstNames: "John", LastName: "Doe", Address: testAddress},
				}},
			},
			match:                 "replacementAttorney",
			matchArticleAndType:   "aReplacementAttorney",
			warningTranslationKey: "actorMatchesDifferentActorNameOrAddressWarningConfirmLater",
		},
		"any other actor": {
			donorProvided: &donordata.Provided{
				LpaID:               "lpa-id",
				CertificateProvider: donordata.CertificateProvider{FirstNames: "Jane", LastName: "Doe"},
				IndependentWitness:  donordata.IndependentWitness{FirstNames: "Jane", LastName: "Doe"},
			},
			match:                 "independentWitness",
			matchArticleAndType:   "theIndependentWitness",
			warningTranslationKey: "actorMatchesDifferentActorNameOrAddressWarningConfirmLater",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			query := url.Values{
				"warningFrom": {donor.PathCertificateProviderAddress.Format("lpa-id")},
				"next":        {"/next-page"},
				"actor":       {actor.TypeCertificateProvider.String()},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T(tc.matchArticleAndType).
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
				T(tc.match).
				Return("5").
				Once()
			localizer.EXPECT().
				Format(
					tc.warningTranslationKey,
					map[string]any{"ArticleAndType": "4", "FullName": "Jane Doe", "MatchArticleAndType": "1", "Type": "2", "TypePlural": "3", "Match": "5"}).
				Return("translatedWarning").
				Once()

			appData := appcontext.Data{LpaID: "lpa-id", Localizer: localizer}
			nextAppData := appData
			nextAppData.Page = "/next-page"

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, WarningInterruptionData{
					App:                 nextAppData,
					Provided:            tc.donorProvided,
					CertificateProvider: &tc.donorProvided.CertificateProvider,
					Notifications: []page.Notification{
						{Heading: "pleaseReviewTheInformationYouHaveEntered", BodyHTML: "translatedWarning"},
					},
					PageTitle: "checkYourCertificateProvidersDetails",
					Next:      "/next-page",
				}).
				Return(nil)

			err := WarningInterruption(template.Execute)(appData, w, r, tc.donorProvided)
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
		"warningFrom": {donor.PathEnterPersonToNotify.Format("lpa-id")},
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
			From:      donor.PathEnterPersonToNotify.FormatQuery("lpa-id", url.Values{"id": {testUID.String()}}),
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
		queryUID, actor string
	}{
		"actor not supported": {
			actor: "huh",
		},
		"not a UID": {
			queryUID: "not a UID",
			actor:    actor.TypeAttorney.String(),
		},
		"attorney not found": {
			queryUID: uid.String(),
			actor:    actor.TypeAttorney.String(),
		},
		"replacement attorney not found": {
			queryUID: uid.String(),
			actor:    actor.TypeReplacementAttorney.String(),
		},
		"person to notify not found": {
			queryUID: uid.String(),
			actor:    actor.TypePersonToNotify.String(),
		},
		"no errors": {
			queryUID: testUID.String(),
			actor:    actor.TypePersonToNotify.String(),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			query := url.Values{
				"actor": {tc.actor},
				"id":    {tc.queryUID},
			}

			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/?"+query.Encode(), nil)

			err := WarningInterruption(nil)(testAppData, w, r, &donordata.Provided{
				LpaID: "lpa-id",
				Donor: donordata.Donor{FirstNames: "Jill", LastName: "Doe"},
				Attorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{
						{UID: testUID, FirstNames: "Jill", LastName: "Doe"},
					},
				},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys: []donordata.Attorney{
						{UID: testUID, FirstNames: "Jill", LastName: "Doe"},
					},
				},
				PeopleToNotify: donordata.PeopleToNotify{
					{UID: testUID, FirstNames: "Joan", LastName: "Doe"},
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
	address := place.Address{Line1: "a", Postcode: "b"}

	donor := &donordata.Provided{
		Donor: donordata.Donor{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "c", Postcode: "d"}},
		Attorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "c", LastName: "d", Address: address},
			{FirstNames: "e", LastName: "f", Address: address},
			{FirstNames: "g", LastName: "h"},
		}},
		ReplacementAttorneys: donordata.Attorneys{Attorneys: []donordata.Attorney{
			{FirstNames: "i", LastName: "j"},
			{FirstNames: "k", LastName: "l", Address: address},
			{FirstNames: "m", LastName: "n", Address: address},
		}},
		CertificateProvider: donordata.CertificateProvider{FirstNames: "o", LastName: "p", Address: address},
		PeopleToNotify: donordata.PeopleToNotify{
			{FirstNames: "q", LastName: "r"},
			{FirstNames: "s", LastName: "t"},
		},
		AuthorisedSignatory: donordata.AuthorisedSignatory{FirstNames: "a", LastName: "s"},
		IndependentWitness:  donordata.IndependentWitness{FirstNames: "i", LastName: "w"},
	}

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "x", "y"))

	assert.Equal(t, actor.TypeDonor, certificateProviderMatches(donor, "a", "b"))

	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(donor, "", "d"))
	assert.Equal(t, actor.TypeAttorney, certificateProviderMatches(donor, "", "F"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "", "h"))

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "", "J"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(donor, "", "l"))
	assert.Equal(t, actor.TypeReplacementAttorney, certificateProviderMatches(donor, "", "n"))

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "o", "p"))

	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "q", "r"))
	assert.Equal(t, actor.TypeNone, certificateProviderMatches(donor, "s", "t"))

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
