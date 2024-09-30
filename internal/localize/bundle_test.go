package localize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBundle(t *testing.T) {
	assert := assert.New(t)
	bundle, err := NewBundle("testdata/en.json", "testdata/cy.json")
	assert.Nil(err)

	en := bundle.For(En)
	assert.Equal("A", en.T("a"))
	assert.Equal("key does not exist", en.T("key does not exist"))

	assert.Equal("A person", en.Format("af", map[string]interface{}{"x": "person"}))
	assert.Equal("key does not exist", en.Format("key does not exist", map[string]interface{}{"x": "person"}))

	assert.Equal("1 ONE", en.Count("c", 1))
	assert.Equal("2 OTHER", en.Count("c", 2))
	assert.Equal("key does not exist", en.Count("key does not exist", 3))

	assert.Equal("1 ONE FORMATTED", en.FormatCount("d", 1, map[string]interface{}{"x": "FORMATTED"}))
	assert.Equal("2 OTHER FORMATTED", en.FormatCount("d", 2, map[string]interface{}{"x": "FORMATTED"}))
	assert.Equal("key does not exist", en.FormatCount("key does not exist", 2, map[string]interface{}{"x": "FORMATTED"}))

	assert.Equal("for john’s birthday", en.Format("p", map[string]any{"x": "John"}))
	assert.Equal("for johns’ birthday", en.Format("p", map[string]any{"x": "Johns"}))

	cy := bundle.For(Cy)
	assert.Equal("C", cy.T("a"))

	assert.Equal("C person", cy.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("1 one", cy.Count("c", 1))
	assert.Equal("2 two", cy.Count("c", 2))
	assert.Equal("3 few", cy.Count("c", 3))
	assert.Equal("4 other", cy.Count("c", 4))
	assert.Equal("5 other", cy.Count("c", 5))
	assert.Equal("6 many", cy.Count("c", 6))
	assert.Equal("7 other", cy.Count("c", 7))

	assert.Equal("1 one formatted", cy.FormatCount("d", 1, map[string]interface{}{"x": "formatted"}))
	assert.Equal("2 two formatted", cy.FormatCount("d", 2, map[string]interface{}{"x": "formatted"}))
	assert.Equal("3 few formatted", cy.FormatCount("d", 3, map[string]interface{}{"x": "formatted"}))
	assert.Equal("4 other formatted", cy.FormatCount("d", 4, map[string]interface{}{"x": "formatted"}))
	assert.Equal("5 other formatted", cy.FormatCount("d", 5, map[string]interface{}{"x": "formatted"}))
	assert.Equal("6 many formatted", cy.FormatCount("d", 6, map[string]interface{}{"x": "formatted"}))
	assert.Equal("7 other formatted", cy.FormatCount("d", 7, map[string]interface{}{"x": "formatted"}))

	assert.Equal("for john birthday", cy.Format("p", map[string]any{"x": "John"}))
}

func TestNewBundleWhenBadFormat(t *testing.T) {
	_, err := NewBundle("testdata/bad/en.json")
	assert.NotNil(t, err)

	_, err = NewBundle("testdata/bad/cy.json")
	assert.NotNil(t, err)
}

func TestNewBundleWhenMissingFile(t *testing.T) {
	_, err := NewBundle("testdata/a.json")
	assert.NotNil(t, err)
}

func TestNewBundleWhenOtherLang(t *testing.T) {
	_, err := NewBundle("testdata/zz.json")
	assert.NotNil(t, err)
}

func TestNewBundleWhenMalformed(t *testing.T) {
	_, err := NewBundle("testdata/malformed/en.json")
	assert.NotNil(t, err)
}

func TestNewBundleWithTransKeys(t *testing.T) {
	assert := assert.New(t)
	bundle, _ := NewBundle("testdata/en.json", "testdata/cy.json")

	en := bundle.For(En)
	en.SetShowTranslationKeys(true)

	assert.Equal("{A} [a]", en.T("a"))
	assert.Equal("{key does not exist} [key does not exist]", en.T("key does not exist"))

	assert.Equal("{A person} [af]", en.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("{1 ONE} [c]", en.Count("c", 1))
	assert.Equal("{2 OTHER} [c]", en.Count("c", 2))

	assert.Equal("{1 ONE FORMATTED} [d]", en.FormatCount("d", 1, map[string]interface{}{"x": "FORMATTED"}))
	assert.Equal("{2 OTHER FORMATTED} [d]", en.FormatCount("d", 2, map[string]interface{}{"x": "FORMATTED"}))

	cy := bundle.For(Cy)
	cy.SetShowTranslationKeys(true)

	assert.Equal("{C} [a]", cy.T("a"))

	assert.Equal("{C person} [af]", cy.Format("af", map[string]interface{}{"x": "person"}))

	assert.Equal("{1 one} [c]", cy.Count("c", 1))
	assert.Equal("{2 two} [c]", cy.Count("c", 2))
	assert.Equal("{3 few} [c]", cy.Count("c", 3))
	assert.Equal("{4 other} [c]", cy.Count("c", 4))
	assert.Equal("{5 other} [c]", cy.Count("c", 5))
	assert.Equal("{6 many} [c]", cy.Count("c", 6))
	assert.Equal("{7 other} [c]", cy.Count("c", 7))

	assert.Equal("{1 one formatted} [d]", cy.FormatCount("d", 1, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{2 two formatted} [d]", cy.FormatCount("d", 2, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{3 few formatted} [d]", cy.FormatCount("d", 3, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{4 other formatted} [d]", cy.FormatCount("d", 4, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{5 other formatted} [d]", cy.FormatCount("d", 5, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{6 many formatted} [d]", cy.FormatCount("d", 6, map[string]interface{}{"x": "formatted"}))
	assert.Equal("{7 other formatted} [d]", cy.FormatCount("d", 7, map[string]interface{}{"x": "formatted"}))
}
