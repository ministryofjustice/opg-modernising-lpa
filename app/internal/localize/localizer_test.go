package localize

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBundle(t *testing.T) {
	bundle := NewBundle("testdata/en.json", "testdata/cy.json")

	en := bundle.For("en")
	assert.Equal(t, "A", en.T("a"))
	assert.Equal(t, template.HTML("B"), en.HTML("b"))

	cy := bundle.For("cy")
	assert.Equal(t, "C", cy.T("a"))
	assert.Equal(t, template.HTML("D"), cy.HTML("b"))
}
