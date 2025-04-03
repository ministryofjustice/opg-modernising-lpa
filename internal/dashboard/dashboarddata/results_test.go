package dashboarddata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultsEmpty(t *testing.T) {
	results := Results{}
	assert.True(t, results.Empty())

	results.Donor = append(results.Donor, Actor{})
	assert.False(t, results.Empty())
}
