package page

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTaskList(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	template := &mockTemplate{}
	template.
		On("Func", w, &taskListData{
			Page: taskListPath,
			L:    localizer,
			Lang: En,
			Sections: []taskListSection{
				{
					Heading: "fillInTheLpa",
					Items: []taskListItem{
						{Name: "provideDonorDetails", Path: donorDetailsPath, Completed: true},
						{Name: "chooseYourContactPreferences", Path: howWouldYouLikeToBeContactedPath, Completed: true},
						{Name: "chooseYourAttorneys"},
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

	TaskList(nil, localizer, En, template.Func, nil)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template)
}

func TestGetTaskListWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	localizer := localize.Localizer{}

	logger := &mockLogger{}
	logger.
		On("Print", expectedError)

	template := &mockTemplate{}
	template.
		On("Func", w, mock.Anything).
		Return(expectedError)

	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	TaskList(logger, localizer, En, template.Func, nil)(w, r)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	mock.AssertExpectationsForObjects(t, template, logger)
}
