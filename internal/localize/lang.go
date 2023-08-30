package localize

type Lang int

func (l Lang) String() string {
	if l == Cy {
		return welshAbbreviation
	}

	return englishAbbreviation
}

const (
	En Lang = iota
	Cy
	englishAbbreviation = "en"
	welshAbbreviation   = "cy"
)
