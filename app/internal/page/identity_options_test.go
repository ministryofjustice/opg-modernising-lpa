package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadIdentityOption(t *testing.T) {
	assert.Equal(t, Passport, readIdentityOption("passport"))
	assert.Equal(t, IdentityOptionUnknown, readIdentityOption("what"))
}

func TestIdentityOptionArticleLabel(t *testing.T) {
	assert.Equal(t, "postOfficeEasyID", EasyID.ArticleLabel())
	assert.Equal(t, "", IdentityOptionUnknown.ArticleLabel())
}
