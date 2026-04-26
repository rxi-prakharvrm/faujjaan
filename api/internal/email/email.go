package email

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v3"
)

type Service interface {
	SendVerificationEmail(ctx context.Context, toEmail, token string) error
	SendOTPEmail(ctx context.Context, toEmail, otp string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, otp string) error
}

type resendService struct {
	client    *resend.Client
	fromEmail string
	baseURL   string
}

func NewService(apiKey, fromEmail, baseURL string) Service {
	if apiKey == "" {
		return &DummyService{baseURL: baseURL}
	}
	client := resend.NewClient(apiKey)
	return &resendService{
		client:    client,
		fromEmail: fromEmail,
		baseURL:   baseURL,
	}
}

func (s *resendService) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", s.baseURL, token)
	html := fmt.Sprintf(`
		<h1>Welcome to Vexo!</h1>
		<p>Please click the link below to verify your email address:</p>
		<p><a href="%s">Verify Email</a></p>
		<p>Or paste this link into your browser: <br>%s</p>
	`, verifyURL, verifyURL)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: "Verify your email address",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		fmt.Printf("resend error: %v\n", err)
	}
	return err
}

func (s *resendService) SendOTPEmail(ctx context.Context, toEmail, otp string) error {
	html := fmt.Sprintf(`
		<h1>Your Vexo Login Code</h1>
		<p>Your one-time password (OTP) is:</p>
		<h2>%s</h2>
		<p>This code will expire in 10 minutes. Do not share it with anyone.</p>
	`, otp)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: "Your Login OTP",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		fmt.Printf("resend error: %v\n", err)
	}
	return err
}

func (s *resendService) SendPasswordResetEmail(ctx context.Context, toEmail, otp string) error {
	html := fmt.Sprintf(`
		<h1>Reset Your Password</h1>
		<p>You requested to reset your password. Use the following code to proceed:</p>
		<h2>%s</h2>
		<p>This code will expire in 15 minutes. If you did not request this, please ignore this email.</p>
	`, otp)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: "Password Reset Code",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		fmt.Printf("resend error: %v\n", err)
	}
	return err
}

type DummyService struct {
	baseURL string
}

func (d *DummyService) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/verify?token=%s", d.baseURL, token)
	fmt.Printf("\n--- [DUMMY EMAIL] ---\nTo: %s\nSubject: Verify Email\nLink: %s\n---------------------\n", toEmail, verifyURL)
	return nil
}

func (d *DummyService) SendOTPEmail(ctx context.Context, toEmail, otp string) error {
	fmt.Printf("\n--- [DUMMY EMAIL] ---\nTo: %s\nSubject: Your OTP\nCode: %s\n---------------------\n", toEmail, otp)
	return nil
}

func (d *DummyService) SendPasswordResetEmail(ctx context.Context, toEmail, otp string) error {
	fmt.Printf("\n--- [DUMMY EMAIL] ---\nTo: %s\nSubject: Password Reset\nCode: %s\n---------------------\n", toEmail, otp)
	return nil
}
