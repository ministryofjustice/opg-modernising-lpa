package pay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PayClient interface {
	CreatePayment(body CreatePaymentBody) (CreatePaymentResponse, error)
	GetPayment(paymentId string) (GetPaymentResponse, error)
}

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Client struct {
	BaseURL    string
	ApiKey     string
	HttpClient Doer
}

type GovUKPayTime time.Time

func (g *GovUKPayTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	s = strings.Trim(s, "0Z")

	t, err := time.Parse(time.RFC3339, s+"00Z")
	if err != nil {
		return err
	}

	*g = GovUKPayTime(t)
	return nil
}

func (g *GovUKPayTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*g))
}

func (g *GovUKPayTime) Format(s string) string {
	t := time.Time(*g)
	return t.Format(s)
}

//"2016-01-21T17:15:000Z"

func (c *Client) CreatePayment(body CreatePaymentBody) (CreatePaymentResponse, error) {
	data, _ := json.Marshal(body)
	reader := bytes.NewReader(data)

	req, err := http.NewRequest("POST", c.BaseURL+"/v1/payments", reader)
	if err != nil {
		return CreatePaymentResponse{}, err
	}

	fmt.Println(c.BaseURL + "/v1/payments")

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

	fmt.Println(c.BaseURL + "/v1/payments/" + paymentId)

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
