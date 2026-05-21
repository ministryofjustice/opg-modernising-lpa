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

func TestCySoftMutate(t *testing.T) {
	assert.Equal(t, "", cySoftMutate(""))
	assert.Equal(t, "Barry", cySoftMutate("Parry"))
	assert.Equal(t, "barry", cySoftMutate("parry"))
	assert.Equal(t, "ary", cySoftMutate("Gary"))
	assert.Equal(t, "ary", cySoftMutate("gary"))
	assert.Equal(t, "Ddarry", cySoftMutate("Darry"))
	assert.Equal(t, "ddarry", cySoftMutate("darry"))
	assert.Equal(t, "R", cySoftMutate("R"))
	assert.Equal(t, "R", cySoftMutate("Rh"))
}
