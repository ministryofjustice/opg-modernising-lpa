package pay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type EventClient interface {
	SendPaymentCreated(ctx context.Context, e event.PaymentCreated) error
}

type Client struct {
	logger      Logger
	doer        Doer
	eventClient EventClient
	baseURL     string
	apiKey      string
}

func New(logger Logger, doer Doer, eventClient EventClient, baseURL, apiKey string) *Client {
	return &Client{
		logger:      logger,
		doer:        doer,
		eventClient: eventClient,
		baseURL:     baseURL,
		apiKey:      apiKey,
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

	// TODO: move this to after payment has been confirmed!
	if err := c.eventClient.SendPaymentCreated(ctx, event.PaymentCreated{
		UID:       lpaUID,
		PaymentID: createPaymentResp.PaymentID,
		Amount:    createPaymentResp.Amount,
	}); err != nil {
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
