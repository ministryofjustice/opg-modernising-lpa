package pay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type CreatePaymentBody struct {
	Amount      int    `json:"amount"`
	Reference   string `json:"reference"`
	Description string `json:"description"`
	ReturnUrl   string `json:"return_url"`
	Email       string `json:"email"`
	Language    string `json:"language"`
}

type State struct {
	Status   string `json:"status"`
	Finished bool   `json:"finished"`
}

type Link struct {
	Href   string `json:"href"`
	Method string `json:"method"`
}

type CreatePaymentResponse struct {
	CreatedDate     time.Time       `json:"created_date"`
	State           State           `json:"state"`
	Links           map[string]Link `json:"_links"`
	Amount          int             `json:"amount"`
	Reference       string          `json:"reference"`
	Description     string          `json:"description"`
	ReturnUrl       string          `json:"return_url"`
	PaymentId       string          `json:"payment_id"`
	PaymentProvider string          `json:"payment_provider"`
	ProviderId      string          `json:"provider_id"`
}

func New(baseURL string, httpClient *http.Client) (Client, error) {
	return Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

func (c *Client) CreatePayment(body CreatePaymentBody) (CreatePaymentResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return CreatePaymentResponse{}, err
	}
	reader := bytes.NewReader(data)

	resp, _ := c.httpClient.Post(c.baseURL+"/v1/payments", "application/json", reader)
	defer resp.Body.Close()

	var createPaymentResp CreatePaymentResponse

	if err := json.NewDecoder(resp.Body).Decode(&createPaymentResp); err != nil {
		return CreatePaymentResponse{}, err
	}

	return createPaymentResp, nil
}
