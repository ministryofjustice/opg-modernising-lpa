package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	signedAt := time.Now()
	attorneyUID := actoruid.New()

	testCases := map[string]struct {
		lpa      *lpadata.Lpa
		provided *attorneydata.Provided
		appData  appcontext.Data
		expected func([]taskListItem) []taskListItem
	}{
		"empty": {
			lpa: &lpadata.Lpa{
				LpaID: "lpa-id",
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID: attorneyUID,
					}},
				},
			},
			provided: &attorneydata.Provided{UID: attorneyUID},
			appData:  testAppData,
			expected: func(items []taskListItem) []taskListItem {
				return items
			},
		},
		"donor gave phone number": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:    attorneyUID,
						Mobile: "07777",
					}},
				},
			},
			provided: &attorneydata.Provided{UID: attorneyUID},
			appData:  testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].Path = attorney.PathYourPreferredLanguage

				return items
			},
		},
		"trust corporation": {
			lpa:      &lpadata.Lpa{LpaID: "lpa-id"},
			provided: &attorneydata.Provided{},
			appData:  testTrustCorporationAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].Path = attorney.PathPhoneNumber

				return items
			},
		},
		"trust corporation with two signatories": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
			},
			provided: &attorneydata.Provided{
				WouldLikeSecondSignatory: form.Yes,
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			appData: testTrustCorporationAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = attorney.PathConfirmYourDetails
				items[1].State = task.StateCompleted
				items[2].Name = "signTheLpaSignatory1"
				items[2].Path = attorney.PathRightsAndResponsibilities

				return append(items, taskListItem{
					Name:  "signTheLpaSignatory2",
					Path:  attorney.PathSign,
					Query: "?second",
				})
			},
		},
		"tasks completed not signed": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
			},
			provided: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = attorney.PathConfirmYourDetails
				items[1].State = task.StateCompleted

				return items
			},
		},
		"tasks completed and signed": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
			},
			provided: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = attorney.PathConfirmYourDetails
				items[1].State = task.StateCompleted
				items[2].Path = attorney.PathRightsAndResponsibilities

				return items
			},
		},
		"completed": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
			},
			provided: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
					SignTheLpa:         task.StateCompleted,
				},
			},
			appData: testAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = attorney.PathConfirmYourDetails
				items[1].State = task.StateCompleted
				items[2].State = task.StateCompleted
				items[2].Path = attorney.PathRightsAndResponsibilities

				return items
			},
		},
		"completed replacement": {
			lpa: &lpadata.Lpa{
				LpaID:                            "lpa-id",
				SignedAt:                         time.Now(),
				WitnessedByCertificateProviderAt: time.Now(),
				CertificateProvider:              lpadata.CertificateProvider{SignedAt: &signedAt},
			},
			provided: &attorneydata.Provided{
				Tasks: attorneydata.Tasks{
					ConfirmYourDetails: task.StateCompleted,
					ReadTheLpa:         task.StateCompleted,
					SignTheLpa:         task.StateCompleted,
				},
			},
			appData: testReplacementAppData,
			expected: func(items []taskListItem) []taskListItem {
				items[0].State = task.StateCompleted
				items[0].Path = attorney.PathConfirmYourDetails
				items[1].State = task.StateCompleted
				items[2].State = task.StateCompleted
				items[2].Path = attorney.PathRightsAndResponsibilities

				return items
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &taskListData{
					App:      tc.appData,
					Lpa:      tc.lpa,
					Provided: tc.provided,
					Items: tc.expected([]taskListItem{
						{Name: "confirmYourDetails", Path: attorney.PathPhoneNumber},
						{Name: "readTheLpa", Path: attorney.PathReadTheLpa},
						{Name: "signTheLpa", Path: attorney.PathRightsAndResponsibilities},
					}),
				}).
				Return(nil)

			err := TaskList(template.Execute)(tc.appData, w, r, tc.provided, tc.lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, mock.Anything).
		Return(expectedError)

	err := TaskList(template.Execute)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{LpaID: "lpa-id"})
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
