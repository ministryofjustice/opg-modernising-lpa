package page

import (
	"fmt"
	"net/http"
)

func IdentityWithEasyIDCallback(yotiClient yotiClient) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		user, err := yotiClient.User(r.FormValue("token"))
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(w, "<!doctype html><p>Hi %s</p>", user.FullName)
		return err
	}
}
