package attorney

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		attorney        *actor.AttorneyProvidedDetails
		attorneys       []*actor.AttorneyProvidedDetails
		signed          bool
		attorneysSigned bool
	}{
		"unsigned": {
			attorney: &actor.AttorneyProvidedDetails{},
		},
		"attorney signed": {
			attorney: &actor.AttorneyProvidedDetails{
				LpaSignedAt: lpaSignedAt,
				Confirmed:   attorneySignedAt,
			},
			attorneys: []*actor.AttorneyProvidedDetails{{}},
			signed:    true,
		},
		"all signed": {
			attorney: &actor.AttorneyProvidedDetails{},
			attorneys: []*actor.AttorneyProvidedDetails{{
				LpaSignedAt: lpaSignedAt,
				Confirmed:   attorneySignedAt,
			}},
			attorneysSigned: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpa := &actor.Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{}}},
			}

			donorStore := newMockDonorStore(t)
			donorStore.
				On("GetAny", r.Context()).
				Return(lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.
				On("GetAny", r.Context()).
				Return(tc.attorneys, nil)

			template := newMockTemplate(t)
			template.
				On("Execute", w, &progressData{App: testAppData, Lpa: lpa, Signed: tc.signed, AttorneysSigned: tc.attorneysSigned}).
				Return(nil)

			err := Progress(template.Execute, attorneyStore, donorStore)(testAppData, w, r, tc.attorney)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestProgressWhenAttorneyStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.Lpa{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAny", r.Context()).
		Return(nil, expectedError)

	err := Progress(nil, attorneyStore, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestProgressWhenDonorStoreErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	lpa := &actor.Lpa{}

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(lpa, expectedError)

	err := Progress(nil, nil, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}

func TestProgressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	donorStore := newMockDonorStore(t)
	donorStore.
		On("GetAny", r.Context()).
		Return(&actor.Lpa{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.
		On("GetAny", r.Context()).
		Return([]*actor.AttorneyProvidedDetails{}, nil)

	template := newMockTemplate(t)
	template.
		On("Execute", w, &progressData{App: testAppData, Lpa: &actor.Lpa{}}).
		Return(expectedError)

	err := Progress(template.Execute, attorneyStore, donorStore)(testAppData, w, r, &actor.AttorneyProvidedDetails{})
	assert.Equal(t, expectedError, err)
}
