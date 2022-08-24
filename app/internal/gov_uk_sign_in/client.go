package govuksignin

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	httpClient       *http.Client
	baseURL          string
	DiscoverData     DiscoverResponse
	AuthCallbackPath string
}

type DiscoverResponse struct {
	AuthorizationEndpoint                      url.URL  `json:"authorization_endpoint"`
	TokenEndpoint                              url.URL  `json:"token_endpoint"`
	RegistrationEndpoint                       string   `json:"registration_endpoint"`
	Issuer                                     string   `json:"issuer"`
	JwksUri                                    string   `json:"jwks_uri"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported"`
	ServiceDocumentation                       string   `json:"service_documentation"`
	RequestUriParameterSupported               bool     `json:"request_uri_parameter_supported"`
	Trustmarks                                 string   `json:"trustmarks"`
	SubjectTypesSupported                      []string `json:"subject_types_supported"`
	UserinfoEndpoint                           url.URL  `json:"userinfo_endpoint"`
	EndSessionEndpoint                         string   `json:"end_session_endpoint"`
	IdTokenSigningAlgValuesSupported           []string `json:"id_token_signing_alg_values_supported"`
	ClaimTypesSupported                        []string `json:"claim_types_supported"`
	ClaimsSupported                            []string `json:"claims_supported"`
	BackchannelLogoutSupported                 bool     `json:"backchannel_logout_supported"`
	BackchannelLogoutSessionSupported          bool     `json:"backchannel_logout_session_supported"`
}

// UnmarshalJSON Allows for unmarshalling to url.URL
func (dr *DiscoverResponse) UnmarshalJSON(data []byte) error {
	type DiscoverResponseClone DiscoverResponse

	tmp := struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		UserinfoEndpoint      string `json:"userinfo_endpoint"`
		*DiscoverResponseClone
	}{
		DiscoverResponseClone: (*DiscoverResponseClone)(dr),
	}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	authEndpointURL, err := url.Parse(tmp.AuthorizationEndpoint)
	if err != nil {
		return err
	}

	tokenEndpointURL, err := url.Parse(tmp.TokenEndpoint)
	if err != nil {
		return err
	}

	userInfoURL, err := url.Parse(tmp.UserinfoEndpoint)
	if err != nil {
		return err
	}

	dr.AuthorizationEndpoint = *authEndpointURL
	dr.TokenEndpoint = *tokenEndpointURL
	dr.UserinfoEndpoint = *userInfoURL

	return nil
}

func NewClient(httpClient *http.Client, baseURL, authCallbackPath string) *Client {
	client := &Client{
		httpClient:       httpClient,
		baseURL:          baseURL,
		AuthCallbackPath: authCallbackPath,
	}

	return client
}

func (c *Client) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) Init() error {
	dr, err := c.DiscoverEndpoints()
	if err != nil {
		return err
	}

	c.DiscoverData = dr
	return c.assertEndpointsHostsMatchIssuerHost()
}
