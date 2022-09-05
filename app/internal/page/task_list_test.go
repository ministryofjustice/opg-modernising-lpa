package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &taskListData{
			App: appData,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{Name: "provideDonorDetails", Path: donorDetailsPath},
						{Name: "chooseYourContactPreferences", Path: howWouldYouLikeToBeContactedPath},
						{Name: "chooseYourAttorneys", Path: chooseAttorneysPath},
						{Name: "chooseYourReplacementAttorneys"},
						{Name: "chooseWhenTheLpaCanBeUsed"},
						{Name: "addRestrictionsToTheLpa"},
						{Name: "chooseYourCertificateProvider"},
						{Name: "checkAndSendToYourCertificateProvider"},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{Name: "payForTheLpa"},
					},
				},
				{
					Heading: "confirmYourIdentity",
					Items: []taskListItem{
						{Name: "confirmYourIdentity"},
					},
				},
				{
					Heading: "signAndRegisterTheLpa",
					Items: []taskListItem{
						{Name: "signTheLpa", Disabled: true},
					},
				},
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetTaskListWhenComplete(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{
		data: Lpa{
			Donor: Donor{
				Address: Address{
					Line1: "this",
				},
			},
			Attorney: Attorney{
				Address: Address{
					Line1: "this",
				},
			},
			Contact: []string{"this"},
		},
	}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, &taskListData{
			App: appData,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{Name: "provideDonorDetails", Path: donorDetailsPath, Completed: true},
						{Name: "chooseYourContactPreferences", Path: howWouldYouLikeToBeContactedPath, Completed: true},
						{Name: "chooseYourAttorneys", Path: chooseAttorneysPath, Completed: true},
						{Name: "chooseYourReplacementAttorneys"},
						{Name: "chooseWhenTheLpaCanBeUsed"},
						{Name: "addRestrictionsToTheLpa"},
						{Name: "chooseYourCertificateProvider"},
						{Name: "checkAndSendToYourCertificateProvider"},
					},
				},
				{
					Heading: "payForTheLpa",
					Items: []taskListItem{
						{Name: "payForTheLpa"},
					},
				},
				{
					Heading: "confirmYourIdentity",
					Items: []taskListItem{
						{Name: "confirmYourIdentity"},
					},
				},
				{
					Heading: "signAndRegisterTheLpa",
					Items: []taskListItem{
						{Name: "signTheLpa", Disabled: true},
					},
				},
			},
		}).
		Return(nil)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()

	dataStore := &mockDataStore{}
	dataStore.
		On("Get", mock.Anything, "session-id").
		Return(nil)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	err := TaskList(template.Func, dataStore)(appData, w, r)
	resp := w.Result()

	assert.Equal(t, expectedError, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, dataStore)
}
