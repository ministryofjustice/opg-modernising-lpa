package localize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLowerFirst(t *testing.T) {
	assert.Equal(t, "hELLO", LowerFirst("HELLO"))
	assert.Equal(t, "hello", LowerFirst("hello"))
}

func TestEnPossessive(t *testing.T) {
	assert.Equal(t, "Barry’s", enPossessive("Barry"))
	assert.Equal(t, "James’", enPossessive("James"))
}

func TestCyAac(t *testing.T) {
	assert.Equal(t, "a Barry", cyAac("Barry"))
	assert.Equal(t, "ac Alan", cyAac("Alan"))
}
