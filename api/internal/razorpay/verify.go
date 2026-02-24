package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func VerifyWebhookSignature(body []byte, signatureHeader string, webhookSecret string) bool {
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signatureHeader))
}

// VerifyPaymentSignature verifies the signature returned by Razorpay Checkout.
// See: https://razorpay.com/docs/payments/server-integration/go/payment-gateway/build-integration/#step-4-verify-payment-signature
func VerifyPaymentSignature(razorpayOrderID, razorpayPaymentID, razorpaySignature, keySecret string) bool {
	msg := razorpayOrderID + "|" + razorpayPaymentID
	mac := hmac.New(sha256.New, []byte(keySecret))
	_, _ = mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(razorpaySignature))
}

