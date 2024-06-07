package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		attorney        *actor.AttorneyProvidedDetails
		attorneys       []lpastore.Attorney
		signed          bool
		attorneysSigned bool
	}{
		"unsigned": {
			attorney: &actor.AttorneyProvidedDetails{},
		},
		"attorney signed": {
			attorney: &actor.AttorneyProvidedDetails{
				SignedAt: attorneySignedAt,
			},
			attorneys: []lpastore.Attorney{{}},
			signed:    true,
		},
		"all signed": {
			attorney: &actor.AttorneyProvidedDetails{},
			attorneys: []lpastore.Attorney{{
				SignedAt: attorneySignedAt,
			}},
			attorneysSigned: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			donor := &lpastore.Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: lpastore.Attorneys{Attorneys: tc.attorneys},
			}

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Get(r.Context()).
				Return(donor, nil)

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &progressData{App: testAppData, Lpa: donor, Signed: tc.signed, AttorneysSigned: tc.attorneysSigned}).
				Return(nil)

			err := Progress(template.Execute, lpaStoreResolvingService)(testAppData, w, r, tc.attorney)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestProgressWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donor := &lpastore.Lpa{}

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(donor, expectedError)

	err := Progress(nil, lpaStoreResolvingService)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestProgressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Get(r.Context()).
		Return(&lpastore.Lpa{}, nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &progressData{App: testAppData, Lpa: &lpastore.Lpa{}}).
		Return(expectedError)

	err := Progress(template.Execute, lpaStoreResolvingService)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}
