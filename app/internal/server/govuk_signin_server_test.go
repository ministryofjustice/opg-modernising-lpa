package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zitadel/oidc/pkg/client/rp"
)

func TestExchangeCodeForToken(t *testing.T) {
	t.Run("POSTs code in request to /token endpoint", func(t *testing.T) {
		code := "code-value"
		state := "state-value"

		url := fmt.Sprintf(
			"/authorize/callback?code=%s&state=%s",
			code,
			state,
		)

		request, _ := http.NewRequest(http.MethodGet, url, nil)
		response := httptest.NewRecorder()

		OIDCServer(response, request)

		options := []rp.Option{
			rp.WithCookieHandler(cookieHandler),
			rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		}

		got := response.Body.String()
		want := "20"

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
