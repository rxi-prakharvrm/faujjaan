package httpapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"clothes-shop/api/internal/auth"
)

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// --- Registration ---

type registerStartRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleRegisterStart(w http.ResponseWriter, r *http.Request) {
	var req registerStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}

	exists, _, err := s.store.CheckEmailExists(r.Context(), req.Email, "customer")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	u, err := s.store.CreateUserWithRole(r.Context(), req.Email, hash, "customer")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	rawToken, _ := generateRandomToken()
	_ = s.store.CreateEmailVerification(r.Context(), u.ID, hashString(rawToken), 24*time.Hour)
	_ = s.email.SendVerificationEmail(r.Context(), u.Email, rawToken)

	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "message": "verification email sent"})
}

type registerVerifyRequest struct {
	Token string `json:"token"`
}

func (s *Server) handleRegisterVerify(w http.ResponseWriter, r *http.Request) {
	var req registerVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	userID, err := s.store.VerifyEmail(r.Context(), hashString(req.Token))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	token, _ := s.auth.IssueToken(userID, auth.RoleCustomer)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user_id": userID, "role": "customer"})
}

// --- Login ---

type loginOptionsRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleLoginOptions(w http.ResponseWriter, r *http.Request) {
	var req loginOptionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	exists, isVerified, err := s.store.CheckEmailExists(r.Context(), req.Email, "customer")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"exists":      exists,
		"is_verified": isVerified,
	})
}

type loginPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLoginPassword(w http.ResponseWriter, r *http.Request) {
	var req loginPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	u, err := s.store.GetUserByEmail(r.Context(), req.Email, "customer")
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if !u.IsVerified {
		writeError(w, http.StatusForbidden, "email not verified")
		return
	}

	if !auth.CheckPasswordHash(req.Password, u.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, _ := s.auth.IssueToken(u.ID, auth.RoleCustomer)
	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": u.ID, "email": u.Email},
	})
}

type otpRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleOTPRequest(w http.ResponseWriter, r *http.Request) {
	var req otpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	exists, _, _ := s.store.CheckEmailExists(r.Context(), req.Email, "customer")
	if !exists {
		writeError(w, http.StatusNotFound, "email not registered")
		return
	}

	otp, _ := generateOTP()
	_ = s.store.CreateOTP(r.Context(), req.Email, hashString(otp), 10*time.Minute)
	_ = s.email.SendOTPEmail(r.Context(), req.Email, otp)

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "otp sent"})
}

type otpVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (s *Server) handleOTPVerify(w http.ResponseWriter, r *http.Request) {
	var req otpVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	ok, err := s.store.VerifyOTP(r.Context(), req.Email, hashString(req.OTP))
	if err != nil || !ok {
		writeError(w, http.StatusUnauthorized, "invalid or expired otp")
		return
	}

	u, _ := s.store.GetUserByEmail(r.Context(), req.Email, "customer")
	token, _ := s.auth.IssueToken(u.ID, auth.RoleCustomer)

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  map[string]any{"id": u.ID, "email": u.Email},
	})
}

// --- Forgot Password ---

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleForgotPasswordRequest(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	exists, _, _ := s.store.CheckEmailExists(r.Context(), req.Email, "customer")
	if !exists {
		// To prevent email enumeration, we should ideally return OK anyway,
		// but since this is a dev project, 404 is cleaner for the user.
		writeError(w, http.StatusNotFound, "email not registered")
		return
	}

	otp, _ := generateOTP()
	err := s.store.CreatePasswordReset(r.Context(), req.Email, hashString(otp), 15*time.Minute)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to store reset request")
		return
	}

	err = s.email.SendPasswordResetEmail(r.Context(), req.Email, otp)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to send reset email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "reset code sent to email"})
}

type forgotPasswordResetRequest struct {
	Email    string `json:"email"`
	OTP      string `json:"otp"`
	Password string `json:"new_password"`
}

func (s *Server) handleForgotPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Email == "" || req.OTP == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email, otp, and new_password are required")
		return
	}

	ok, err := s.store.VerifyPasswordReset(r.Context(), req.Email, hashString(req.OTP))
	if err != nil || !ok {
		msg := "invalid or expired code"
		if err != nil {
			msg = err.Error()
		}
		writeError(w, http.StatusUnauthorized, msg)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	if err := s.store.UpdatePassword(r.Context(), req.Email, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": "password updated successfully"})
}
