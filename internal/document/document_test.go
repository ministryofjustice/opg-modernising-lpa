package document

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDocumentsDelete(t *testing.T) {
	documents := Documents{
		{Key: "a-key"},
		{Key: "another-key"},
	}

	assert.True(t, documents.Delete("a-key"))
	assert.Equal(t, Documents{{Key: "another-key"}}, documents)

	assert.True(t, documents.Delete("another-key"))
	assert.Equal(t, Documents{}, documents)

	assert.False(t, documents.Delete("not-a-key"))
}

func TestDocumentsKeys(t *testing.T) {
	documents := Documents{
		{Key: "a-key"},
		{Key: "another-key"},
	}

	assert.Equal(t, []string{"a-key", "another-key"}, documents.Keys())
}

func TestDocumentsGet(t *testing.T) {
	documents := Documents{
		{Key: "a-key"},
		{Key: "another-key"},
	}

	assert.Equal(t, Document{Key: "a-key"}, documents.Get("a-key"))
	assert.Equal(t, Document{Key: "another-key"}, documents.Get("another-key"))
	assert.Equal(t, Document{}, documents.Get("not-a-key"))
}

func TestDocumentsPut(t *testing.T) {
	documents := Documents{
		{Key: "a-key", Filename: "a-filename"},
		{Key: "another-key", Filename: "another-filename"},
	}

	documents.Put(Document{Key: "a-key", Filename: "a-new-filename"})
	assert.Equal(t, Document{Key: "a-key", Filename: "a-new-filename"}, documents[0])

	documents.Put(Document{Key: "new-key", Filename: "a-filename"})
	assert.Equal(t, Document{Key: "new-key", Filename: "a-filename"}, documents[2])
}

func TestDocumentsInfectedFilenames(t *testing.T) {
	documents := Documents{
		{Key: "a-key", Filename: "a-filename"},
		{Key: "another-key", Filename: "another-filename", VirusDetected: true},
	}

	assert.Equal(t, []string{"another-filename"}, documents.InfectedFilenames())
}

func TestDocumentsScanned(t *testing.T) {
	documents := Documents{
		{Key: "a-key", Filename: "a-filename"},
		{Key: "another-key", Filename: "another-filename", Scanned: true},
		{Key: "more-key", Filename: "another-filename", Scanned: true},
	}

	assert.Equal(t, Documents{
		{Key: "another-key", Filename: "another-filename", Scanned: true},
		{Key: "more-key", Filename: "another-filename", Scanned: true},
	}, documents.Scanned())
}

func TestDocumentsNotScanned(t *testing.T) {
	documents := Documents{
		{Key: "a-key", Filename: "a-filename", Scanned: true},
		{Key: "another-key", Filename: "another-filename"},
		{Key: "more-key", Filename: "another-filename"},
	}

	assert.Equal(t, Documents{
		{Key: "another-key", Filename: "another-filename"},
		{Key: "more-key", Filename: "another-filename"},
	}, documents.NotScanned())
}

func TestDocumentsFilenames(t *testing.T) {
	documents := Documents{
		{Key: "a-key", Filename: "a-filename"},
		{Key: "another-key", Filename: "another-filename", VirusDetected: true},
	}

	assert.Equal(t, []string{"a-filename", "another-filename"}, documents.Filenames())
}

func TestDocumentsSent(t *testing.T) {
	now := time.Now()

	documents := Documents{
		{Key: "a-key", Filename: "a-filename"},
		{Key: "another-key", Filename: "another-filename", Scanned: true},
		{Key: "more-key", Filename: "more-filename", Sent: now},
		{Key: "further-key", Filename: "further-filename", Scanned: true},
	}

	assert.Equal(t, Documents{
		{Key: "more-key", Filename: "more-filename", Sent: now},
	}, documents.Sent())
}

func TestDocumentsScannedNotSent(t *testing.T) {
	now := time.Now()

	documents := Documents{
		{Key: "a-key", Filename: "a-filename"},
		{Key: "another-key", Filename: "another-filename", Scanned: true},
		{Key: "more-key", Filename: "more-filename", Sent: now},
		{Key: "further-key", Filename: "further-filename", Scanned: true},
	}

	assert.Equal(t, Documents{
		{Key: "another-key", Filename: "another-filename", Scanned: true},
		{Key: "further-key", Filename: "further-filename", Scanned: true},
	}, documents.ScannedNotSent())
}
