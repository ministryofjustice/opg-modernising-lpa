package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLpaPath(t *testing.T) {
	testCases := map[string]struct {
		url               string
		expectedIsLpaPage bool
	}{
		"dashboard": {
			url:               Paths.Dashboard + "?someQuery=5",
			expectedIsLpaPage: false,
		},
		"start": {
			url:               Paths.Start + "?someQuery=6",
			expectedIsLpaPage: false,
		},
		"any other page": {
			url:               "/other?someQuery=7",
			expectedIsLpaPage: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsLpaPage, IsLpaPath(tc.url))
		})
	}
}
