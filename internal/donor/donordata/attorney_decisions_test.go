package donordata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/stretchr/testify/assert"
)

func TestMakeAttorneyDecisions(t *testing.T) {
	testcases := map[string]struct {
		existing AttorneyDecisions
		how      lpadata.AttorneysAct
		details  string
		expected AttorneyDecisions
	}{
		"without details": {
			existing: AttorneyDecisions{},
			how:      lpadata.Jointly,
			details:  "hey",
			expected: AttorneyDecisions{How: lpadata.Jointly},
		},
		"with details": {
			existing: AttorneyDecisions{},
			how:      lpadata.JointlyForSomeSeverallyForOthers,
			details:  "hey",
			expected: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "hey"},
		},
		"same how without details": {
			existing: AttorneyDecisions{How: lpadata.Jointly},
			how:      lpadata.Jointly,
			details:  "hey",
			expected: AttorneyDecisions{How: lpadata.Jointly},
		},
		"same how with details": {
			existing: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "what"},
			how:      lpadata.JointlyForSomeSeverallyForOthers,
			details:  "hey",
			expected: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers, Details: "hey"},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, MakeAttorneyDecisions(tc.existing, tc.how, tc.details))
		})
	}
}

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
