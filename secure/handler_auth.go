package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/flutterize/backend/config"
	"github.com/flutterize/backend/model"
	"github.com/flutterize/backend/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
	cfg      *config.Config
}

func NewAuthHandler(userRepo *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, cfg: cfg}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Identifier == "" || req.Password == "" {
		http.Error(w, `{"error":"identifier and password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.FindByEmailOrPhone(req.Identifier)
	if err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.LoginResponse{
		Token: tokenString,
		User:  *user,
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Phone == "" || req.Fullname == "" || req.Birthdate == "" || req.Password == "" {
		http.Error(w, `{"error":"all fields are required: email, phone, fullname, birthdate, password"}`, http.StatusBadRequest)
		return
	}

	// SECURE #5: Generic error message prevents user enumeration
	existing, _ := h.userRepo.FindByEmailOrPhone(req.Email)
	if existing != nil {
		http.Error(w, `{"error":"registration failed"}`, http.StatusConflict)
		return
	}
	existing, _ = h.userRepo.FindByEmailOrPhone(req.Phone)
	if existing != nil {
		http.Error(w, `{"error":"registration failed"}`, http.StatusConflict)
		return
	}

	// SECURE #7: Sanitize HTML in fullname to prevent stored XSS
	req.Fullname = sanitizeHTML(req.Fullname)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.Create(req, string(hashedPassword))
	if err != nil {
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "registration successful",
		"user":    user,
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Identifier == "" {
		http.Error(w, `{"error":"identifier is required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.FindByEmailOrPhone(req.Identifier)
	if err != nil {
		// Return success even if user not found to prevent enumeration
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "if the account exists, a reset token has been sent"})
		return
	}

	// Generate reset token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	resetToken := hex.EncodeToString(tokenBytes)
	expiry := time.Now().Add(15 * time.Minute)

	if err := h.userRepo.SetResetToken(user.ID, resetToken, expiry); err != nil {
		http.Error(w, `{"error":"failed to generate reset token"}`, http.StatusInternalServerError)
		return
	}

	// Send email via SMTP (MockMail)
	go h.sendResetEmail(user.Email, resetToken)

	// Send SMS via MockSMS
	go h.sendResetSMS(user.Phone, resetToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "if the account exists, a reset token has been sent",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"token and new_password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.FindByResetToken(req.Token)
	if err != nil {
		http.Error(w, `{"error":"invalid or expired reset token"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
		http.Error(w, `{"error":"failed to update password"}`, http.StatusInternalServerError)
		return
	}

	h.userRepo.ClearResetToken(user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "password reset successful"})
}

func (h *AuthHandler) sendResetEmail(email, token string) {
	smtpAddr := fmt.Sprintf("%s:%s", h.cfg.SMTPHost, h.cfg.SMTPPort)

	from := "noreply@flutterize.lab"
	subject := "Password Reset - Flutterize"
	body := fmt.Sprintf("Your password reset token is: %s\n\nThis token expires in 15 minutes.", token)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, email, subject, body)

	err := smtp.SendMail(smtpAddr, nil, from, []string{email}, []byte(msg))
	if err != nil {
		log.Printf("Failed to send reset email to %s: %v", email, err)
	} else {
		log.Printf("Reset email sent to %s", email)
	}
}

func (h *AuthHandler) sendResetSMS(phone, token string) {
	smsURL := fmt.Sprintf("%s/send", h.cfg.SMSGatewayURL)

	payload := map[string]string{
		"phone":   phone,
		"message": fmt.Sprintf("Your Flutterize password reset token is: %s", token),
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(smsURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Failed to send reset SMS to %s: %v", phone, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("Reset SMS sent to %s", phone)
}

// sanitizeHTML strips dangerous HTML tags from input to prevent stored XSS
func sanitizeHTML(input string) string {
	// Remove script tags and their content
	for {
		lower := strings.ToLower(input)
		start := strings.Index(lower, "<script")
		if start == -1 {
			break
		}
		end := strings.Index(lower[start:], "</script>")
		if end == -1 {
			input = input[:start]
		} else {
			input = input[:start] + input[start+end+9:]
		}
	}
	// Remove remaining HTML tags
	var result strings.Builder
	inTag := false
	for _, ch := range input {
		if ch == '<' {
			inTag = true
			continue
		}
		if ch == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
