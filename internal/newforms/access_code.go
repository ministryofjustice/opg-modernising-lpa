package newforms

import "net/http"

type AccessCodeForm struct {
	Form
	DonorLastName *String
	AccessCode    *String
}

func NewAccessCodeForm(l Localizer) *AccessCodeForm {
	return &AccessCodeForm{
		DonorLastName: NewString("donor-last-name", l.T("donorLastName")).
			NotEmpty().
			MaxLength(61),
		AccessCode: NewString("access-code", l.T("yourAccessCode")).
			Replace(
				" ", "",
				"-", "",
			).
			NotEmpty().
			Length(8, CustomError(l.T("theAccessCodeYouEnter"))),
	}
}

func (f *AccessCodeForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r,
		f.DonorLastName,
		f.AccessCode,
	)
}
