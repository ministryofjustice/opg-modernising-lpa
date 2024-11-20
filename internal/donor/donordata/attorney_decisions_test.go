package donordata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestAttorneyDecisionsIsComplete(t *testing.T) {
	testcases := map[string]struct {
		decisions AttorneyDecisions
		expected  bool
	}{
		"how set": {
			decisions: AttorneyDecisions{How: lpadata.Jointly},
			expected:  true,
		},
		"missing how": {
			decisions: AttorneyDecisions{},
			expected:  false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.decisions.IsComplete())
		})
	}
}
