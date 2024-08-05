package document

import (
	"slices"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

type Document struct {
	PK            dynamo.LpaKeyType
	SK            dynamo.DocumentKeyType
	Filename      string
	VirusDetected bool
	Scanned       bool
	Key           string
	Uploaded      time.Time
	Sent          time.Time
}

type Documents []Document

func (ds *Documents) Delete(documentKey string) bool {
	idx := slices.IndexFunc(*ds, func(ds Document) bool { return ds.Key == documentKey })
	if idx == -1 {
		return false
	}

	*ds = slices.Delete(*ds, idx, idx+1)

	return true
}

func (ds *Documents) Keys() []string {
	var keys []string

	for _, ds := range *ds {
		keys = append(keys, ds.Key)
	}

	return keys
}

func (ds *Documents) Put(scannedDocument Document) {
	idx := slices.IndexFunc(*ds, func(ds Document) bool { return ds.Key == scannedDocument.Key })
	if idx == -1 {
		*ds = append(*ds, scannedDocument)
	} else {
		(*ds)[idx] = scannedDocument
	}
}

func (ds *Documents) InfectedFilenames() []string {
	var filenames []string

	for _, d := range *ds {
		if d.VirusDetected {
			filenames = append(filenames, d.Filename)
		}
	}

	return filenames
}

func (ds *Documents) Scanned() Documents {
	var documents Documents

	for _, d := range *ds {
		if d.Scanned {
			documents = append(documents, d)
		}
	}

	return documents
}

func (ds *Documents) NotScanned() Documents {
	var documents Documents

	for _, d := range *ds {
		if !d.Scanned {
			documents = append(documents, d)
		}
	}

	return documents
}

func (ds *Documents) Filenames() []string {
	var filenames []string

	for _, ds := range *ds {
		filenames = append(filenames, ds.Filename)
	}

	return filenames
}

func (ds *Documents) Get(documentKey string) Document {
	for _, d := range *ds {
		if d.Key == documentKey {
			return d
		}
	}

	return Document{}
}

func (ds *Documents) Sent() Documents {
	var documents Documents

	for _, d := range *ds {
		if !d.Sent.IsZero() {
			documents = append(documents, d)
		}
	}

	return documents
}

func (ds *Documents) ScannedNotSent() Documents {
	var documents Documents

	for _, d := range *ds {
		if d.Sent.IsZero() && d.Scanned {
			documents = append(documents, d)
		}
	}

	return documents
}
