package pay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
)

var paymentsURLRe = regexp.MustCompile(`^https://[a-z]+\.payments\.service\.gov\.uk/.+$`)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type Client struct {
	logger      Logger
	doer        Doer
	baseURL     string
	apiKey      string
	canRedirect bool
}

func New(logger Logger, doer Doer, baseURL, apiKey string, canRedirect bool) *Client {
	return &Client{
		logger:      logger,
		doer:        doer,
		baseURL:     baseURL,
		apiKey:      apiKey,
		canRedirect: canRedirect,
	}
}

func (c *Client) CreatePayment(ctx context.Context, lpaUID string, body CreatePaymentBody) (*CreatePaymentResponse, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/payments", &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		c.logger.ErrorContext(ctx, "payment failed",
			slog.String("body", string(data)),
			slog.Int("statusCode", resp.StatusCode))

		return nil, fmt.Errorf("expected 201 got %d", resp.StatusCode)
	}

	var createPaymentResp CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&createPaymentResp); err != nil {
		return nil, err
	}

	return &createPaymentResp, nil
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (GetPaymentResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/payments/"+paymentID, nil)
	if err != nil {
		return GetPaymentResponse{}, err
	}

	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.doer.Do(req)
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

func (c *Client) CanRedirect(url string) bool {
	return c.canRedirect && paymentsURLRe.MatchString(url)
}
