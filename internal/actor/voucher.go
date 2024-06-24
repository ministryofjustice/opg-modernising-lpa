package actor

type Voucher struct {
	FirstNames string
	LastName   string
	Email      string
}

func (v Voucher) FullName() string {
	return v.FirstNames + " " + v.LastName
}
