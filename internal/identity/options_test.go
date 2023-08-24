package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadOption(t *testing.T) {
	assert.Equal(t, Passport, ReadOption("passport"))
	assert.Equal(t, UnknownOption, ReadOption("what"))
}

func TestOptionArticleLabel(t *testing.T) {
	assert.Equal(t, "postOfficeEasyID", EasyID.ArticleLabel())
	assert.Equal(t, "", UnknownOption.ArticleLabel())
}
