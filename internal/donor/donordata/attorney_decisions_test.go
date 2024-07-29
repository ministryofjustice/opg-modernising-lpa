package donordata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeAttorneyDecisions(t *testing.T) {
	testcases := map[string]struct {
		existing AttorneyDecisions
		how      AttorneysAct
		details  string
		expected AttorneyDecisions
	}{
		"without details": {
			existing: AttorneyDecisions{},
			how:      Jointly,
			details:  "hey",
			expected: AttorneyDecisions{How: Jointly},
		},
		"with details": {
			existing: AttorneyDecisions{},
			how:      JointlyForSomeSeverallyForOthers,
			details:  "hey",
			expected: AttorneyDecisions{How: JointlyForSomeSeverallyForOthers, Details: "hey"},
		},
		"same how without details": {
			existing: AttorneyDecisions{How: Jointly},
			how:      Jointly,
			details:  "hey",
			expected: AttorneyDecisions{How: Jointly},
		},
		"same how with details": {
			existing: AttorneyDecisions{How: JointlyForSomeSeverallyForOthers, Details: "what"},
			how:      JointlyForSomeSeverallyForOthers,
			details:  "hey",
			expected: AttorneyDecisions{How: JointlyForSomeSeverallyForOthers, Details: "hey"},
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
			decisions: AttorneyDecisions{How: Jointly},
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
