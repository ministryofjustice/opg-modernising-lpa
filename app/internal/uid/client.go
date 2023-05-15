package uid

import (
	"bytes"
	"io"
	"net/http"
)

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl    string
	httpClient Doer
}

func New(baseUrl string, httpClient Doer) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}

func (c *Client) CreateCase(body string) (string, error) {
	r, err := http.NewRequest(http.MethodGet, c.baseUrl+"/cases", bytes.NewReader([]byte(body)))
	if err != nil {
		return "", err
	}

	r.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	return string(b), nil
}
