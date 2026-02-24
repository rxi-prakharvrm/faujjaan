package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"clothes-shop/api/internal/razorpay"
)

type razorpayVerifyRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

func (s *Server) handleRazorpayVerify(w http.ResponseWriter, r *http.Request) {
	if s.cfg.RazorpayKeySecret == "" {
		writeError(w, http.StatusBadRequest, "razorpay not configured")
		return
	}

	var req razorpayVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.RazorpayOrderID == "" || req.RazorpayPaymentID == "" || req.RazorpaySignature == "" {
		writeError(w, http.StatusBadRequest, "missing razorpay fields")
		return
	}
	if !razorpay.VerifyPaymentSignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature, s.cfg.RazorpayKeySecret) {
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	if err := s.store.MarkPaymentAuthorized(r.Context(), req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to persist payment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

type razorpayWebhook struct {
	Event   string `json:"event"`
	Payload struct {
		Payment struct {
			Entity struct {
				ID      string `json:"id"`
				OrderID string `json:"order_id"`
				Status  string `json:"status"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

func (s *Server) handleRazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	if s.cfg.RazorpayWebhookSecret == "" {
		writeError(w, http.StatusBadRequest, "webhook not configured")
		return
	}
	sig := r.Header.Get("X-Razorpay-Signature")
	if sig == "" {
		writeError(w, http.StatusBadRequest, "missing signature header")
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}
	if !razorpay.VerifyWebhookSignature(body, sig, s.cfg.RazorpayWebhookSecret) {
		writeError(w, http.StatusUnauthorized, "invalid webhook signature")
		return
	}

	var evt razorpayWebhook
	if err := json.Unmarshal(body, &evt); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	orderID := evt.Payload.Payment.Entity.OrderID
	paymentID := evt.Payload.Payment.Entity.ID
	if orderID == "" || paymentID == "" {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	switch evt.Event {
	case "payment.authorized":
		_ = s.store.MarkPaymentAuthorized(r.Context(), orderID, paymentID, "")
	case "payment.captured":
		_ = s.store.MarkPaymentCaptured(r.Context(), orderID, paymentID)
	case "payment.failed":
		_ = s.store.MarkPaymentFailed(r.Context(), orderID, paymentID)
	default:
		// ignore
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

