package lpadata

type Correspondent struct {
	FirstNames string
	LastName   string
	Email      string
	Phone      string
}

func (c Correspondent) FullName() string {
	return c.FirstNames + " " + c.LastName
}
