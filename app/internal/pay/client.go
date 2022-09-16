package pay

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type PayClient interface {
	CreatePayment(body CreatePaymentBody) (CreatePaymentResponse, error)
	GetPayment(paymentId string) (GetPaymentResponse, error)
}

type Client struct {
	BaseURL    string
	ApiKey     string
	HttpClient *http.Client
}

func (c *Client) CreatePayment(body CreatePaymentBody) (CreatePaymentResponse, error) {
	data, _ := json.Marshal(body)
	reader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/payments", reader)
	if err != nil {
		return CreatePaymentResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+c.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return CreatePaymentResponse{}, err
	}

	defer resp.Body.Close()

	var createPaymentResp CreatePaymentResponse

	if err := json.NewDecoder(resp.Body).Decode(&createPaymentResp); err != nil {
		return CreatePaymentResponse{}, err
	}

	return createPaymentResp, nil
}

func (c *Client) GetPayment(paymentId string) (GetPaymentResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/v1/payments/"+paymentId, nil)
	if err != nil {
		return GetPaymentResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+c.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)

	defer resp.Body.Close()

	var getPaymentResponse GetPaymentResponse

	if err := json.NewDecoder(resp.Body).Decode(&getPaymentResponse); err != nil {
		return GetPaymentResponse{}, err
	}

	return getPaymentResponse, nil
}
