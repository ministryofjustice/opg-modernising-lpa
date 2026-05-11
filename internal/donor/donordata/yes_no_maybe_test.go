package donordata

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/stretchr/testify/assert"
)

func TestYesNoMaybeForm(t *testing.T) {
	testcases := map[string]struct {
		value YesNoMaybe
		error newforms.Error
	}{
		"yes": {
			value: Yes,
		},
		"no": {
			value: No,
		},
		"maybe": {
			value: Maybe,
		},
		"not selected": {
			error: newforms.NewSelectError("hey"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			form := NewYesNoMaybeForm("hey")

			query := url.Values{
				form.Enum.Name: {tc.value.String()},
			}

			r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(query.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			assert.Equal(t, tc.error == nil, form.Parse(r))
			if tc.error != nil {
				assert.Equal(t, tc.error, form.Enum.Error)
			}
		})
	}
}
