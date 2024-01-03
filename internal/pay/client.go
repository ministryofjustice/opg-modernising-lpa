package pay

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Client struct {
	BaseURL    string
	ApiKey     string
	HttpClient Doer
}

type GovUKPayTime time.Time

func (g *GovUKPayTime) UnmarshalText(b []byte) error {
	t, err := time.Parse(time.RFC3339Nano, string(b))
	if err != nil {
		return err
	}

	*g = GovUKPayTime(t)
	return nil
}

func (g GovUKPayTime) MarshalText() ([]byte, error) {
	return []byte(g.Format(time.RFC3339Nano)), nil
}

func (g GovUKPayTime) Format(s string) string {
	return time.Time(g).Format(s)
}

func (c *Client) CreatePayment(ctx context.Context, body CreatePaymentBody) (CreatePaymentResponse, error) {
	data, _ := json.Marshal(body)
	reader := bytes.NewReader(data)

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/v1/payments", reader)
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

func (c *Client) GetPayment(ctx context.Context, paymentId string) (GetPaymentResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/v1/payments/"+paymentId, nil)
	if err != nil {
		return GetPaymentResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+c.ApiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HttpClient.Do(req)

	if err != nil {
		return GetPaymentResponse{}, err
	}

	defer resp.Body.Close()

	var getPaymentResponse GetPaymentResponse

	if err := json.NewDecoder(resp.Body).Decode(&getPaymentResponse); err != nil {
		return GetPaymentResponse{}, err
	}

	return getPaymentResponse, nil
}
