package localize

type Lang int

func (l Lang) String() string {
	if l == Cy {
		return welshAbbreviation
	}

	return englishAbbreviation
}

func (l Lang) URL(path string) string {
	if l == Cy {
		return "/" + Cy.String() + path
	}

	return path
}

const (
	En Lang = iota
	Cy
	englishAbbreviation = "en"
	welshAbbreviation   = "cy"
)
