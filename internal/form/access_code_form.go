package form

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type AccessCodeForm struct {
	DonorLastName string
	AccessCode    string
	AccessCodeRaw string
	FieldNames    struct {
		DonorLastName string
		AccessCode    string
	}
}

func NewAccessCodeForm() *AccessCodeForm {
	form := &AccessCodeForm{}
	form.FieldNames.DonorLastName = FieldNames.DonorLastName
	form.FieldNames.AccessCode = FieldNames.AccessCode
	return form
}

func (f *AccessCodeForm) Read(r *http.Request) {
	f.DonorLastName = PostFormString(r, "donor-last-name")
	f.AccessCode = postFormAccessCode(r, "access-code")
	f.AccessCodeRaw = PostFormString(r, "access-code")
}

func (f *AccessCodeForm) Validate() validation.List {
	var errors validation.List

	errors.String("donor-last-name", "donorLastName", f.DonorLastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("access-code", "yourAccessCode", f.AccessCode,
		validation.Empty())

	errors.String("access-code", "theAccessCodeYouEnter", f.AccessCode,
		validation.StringLength(8))

	return errors
}

func postFormAccessCode(r *http.Request, name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue(name), " ", ""), "-", "")
}
