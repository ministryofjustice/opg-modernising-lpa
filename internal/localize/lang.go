package localize

//go:generate enumerator -type Lang -linecomment -empty
type Lang byte

func (i Lang) URL(path string) string {
	if i == Cy {
		return "/" + Cy.String() + path
	}

	return path
}

const (
	En Lang = iota + 1 // en
	Cy                 // cy
)
