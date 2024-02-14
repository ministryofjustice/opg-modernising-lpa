package pacttest

import (
	"net/http"
	"testing"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

func TestOnly(t *testing.T) {
	mockProvider, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "modernising-lpa",
		Provider: "data-lpa-store",
		LogDir:   "../../logs",
		PactDir:  "../../pacts",
	})
	if assert.Nil(t, err) {
		return
	}

	mockProvider.
		AddInteraction().
		Given("a b c").
		UponReceiving("x y z").
		WithRequest(http.MethodGet, "/some/path", func(b *consumer.V2RequestBuilder) {
			b.Header("Content-Type", matchers.String("application/json"))
		}).
		WillRespondWith(http.StatusOK)
}
