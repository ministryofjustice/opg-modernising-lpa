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

type InitialReplacementAttorneyEmail struct {
	DonorFullName             string
	LpaType                   string
	AttorneyFullName          string
	DonorFirstNames           string
	AttorneyStartPageURL      string
	ShareCode                 string
	DonorFirstNamesPossessive string
}

func (e InitialReplacementAttorneyEmail) emailID(isProduction bool) string {
	if isProduction {
		return "8d335239-7002-4825-8393-cc00ad246648"
	}

	return "738d500f-b674-4e1e-8039-a7be53fce528"
}
