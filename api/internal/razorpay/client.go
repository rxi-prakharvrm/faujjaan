package razorpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	keyID     string
	keySecret string
	http      *http.Client
}

func NewClient(keyID, keySecret string) *Client {
	return &Client{
		keyID:     keyID,
		keySecret: keySecret,
		http:      &http.Client{Timeout: 15 * time.Second},
	}
}

type createOrderRequest struct {
	Amount         int    `json:"amount"`
	Currency       string `json:"currency"`
	Receipt        string `json:"receipt"`
	PaymentCapture int    `json:"payment_capture"`
}

type createOrderResponse struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
	Status   string `json:"status"`
}

func (c *Client) CreateOrder(ctx context.Context, amount int, currency string, receipt string) (string, error) {
	body, _ := json.Marshal(createOrderRequest{
		Amount:         amount,
		Currency:       currency,
		Receipt:        receipt,
		PaymentCapture: 1,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.razorpay.com/v1/orders", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.keyID, c.keySecret)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("razorpay create order failed: status=%d", res.StatusCode)
	}
	var out createOrderResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", fmt.Errorf("razorpay create order returned empty id")
	}
	return out.ID, nil
}

