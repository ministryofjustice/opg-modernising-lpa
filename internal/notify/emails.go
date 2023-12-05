package notify

type InitialOriginalAttorneyEmail struct {
	DonorFullName             string
	LpaType                   string
	AttorneyFullName          string
	DonorFirstNames           string
	AttorneyStartPageURL      string
	ShareCode                 string
	DonorFirstNamesPossessive string
}

func (e InitialOriginalAttorneyEmail) emailID(isProduction bool) string {
	if isProduction {
		return "080071dc-0434-4b13-adb7-c4e5612c4b47"
	}

	return "376d7ef2-7941-46c2-b372-bacca0e00c1d"
}
