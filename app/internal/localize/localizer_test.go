package localize

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBundle(t *testing.T) {
	assert := assert.New(t)
	bundle := NewBundle("testdata/en.json", "testdata/cy.json")

	en := bundle.For("en")
	assert.Equal("A", en.T("a"))
	assert.Equal("A person", en.Format("af", map[string]interface{}{"x": "person"}))
	assert.Equal(template.HTML("B"), en.HTML("b"))
	assert.Equal("1 ONE", en.Count("c", 1))
	assert.Equal("2 OTHER", en.Count("c", 2))

	cy := bundle.For("cy")
	assert.Equal("C", cy.T("a"))
	assert.Equal("C person", cy.Format("af", map[string]interface{}{"x": "person"}))
	assert.Equal(template.HTML("D"), cy.HTML("b"))
	assert.Equal("1 one", cy.Count("c", 1))
	assert.Equal("2 two", cy.Count("c", 2))
	assert.Equal("3 few", cy.Count("c", 3))
}
