package attorneypage

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	attorneySignedAt := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		provided        *attorneydata.Provided
		attorneys       []lpadata.Attorney
		signed          bool
		attorneysSigned bool
	}{
		"unsigned": {
			provided: &attorneydata.Provided{},
		},
		"attorney signed": {
			provided: &attorneydata.Provided{
				SignedAt: attorneySignedAt,
			},
			attorneys: []lpadata.Attorney{{}},
			signed:    true,
		},
		"all signed": {
			provided: &attorneydata.Provided{},
			attorneys: []lpadata.Attorney{{
				SignedAt: &attorneySignedAt,
			}},
			attorneysSigned: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodGet, "/", nil)

			lpa := &lpadata.Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: lpadata.Attorneys{Attorneys: tc.attorneys},
			}

			template := newMockTemplate(t)
			template.EXPECT().
				Execute(w, &progressData{App: testAppData, Lpa: lpa, Signed: tc.signed, AttorneysSigned: tc.attorneysSigned}).
				Return(nil)

			err := Progress(template.Execute)(testAppData, w, r, tc.provided, lpa)
			resp := w.Result()

			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestProgressWhenTemplateErrors(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	template := newMockTemplate(t)
	template.EXPECT().
		Execute(w, &progressData{App: testAppData, Lpa: &lpadata.Lpa{}}).
		Return(expectedError)

	err := Progress(template.Execute)(testAppData, w, r, &attorneydata.Provided{}, &lpadata.Lpa{})
	assert.Equal(t, expectedError, err)
}
