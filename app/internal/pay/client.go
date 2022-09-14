package pay

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, httpClient *http.Client) (Client, error) {
	return Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}, nil
}

func (c *Client) CreatePayment(body CreatePaymentBody) (CreatePaymentResponse, error) {
	data, _ := json.Marshal(body)
	reader := bytes.NewReader(data)

	resp, _ := c.httpClient.Post(c.baseURL+"/v1/payments", "application/json", reader)
	defer resp.Body.Close()

	var createPaymentResp CreatePaymentResponse

	if err := json.NewDecoder(resp.Body).Decode(&createPaymentResp); err != nil {
		return CreatePaymentResponse{}, err
	}

	return createPaymentResp, nil
}

func (c *Client) GetPayment(paymentId string) (GetPaymentResponse, error) {
	resp, _ := c.httpClient.Get(c.baseURL + "/v1/payments/" + paymentId)
	defer resp.Body.Close()

	var getPaymentResponse GetPaymentResponse

	if err := json.NewDecoder(resp.Body).Decode(&getPaymentResponse); err != nil {
		return GetPaymentResponse{}, err
	}

	return getPaymentResponse, nil
}
