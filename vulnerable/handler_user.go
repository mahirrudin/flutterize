package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/flutterize/backend/middleware"
	"github.com/flutterize/backend/model"
	"github.com/flutterize/backend/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	// VULN #2: IDOR — accepts user-supplied ID to view any profile
	idParam := r.URL.Query().Get("id")
	if idParam != "" {
		parsedID, err := strconv.ParseUint(idParam, 10, 64)
		if err == nil {
			userID = parsedID
		}
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Fullname == "" {
		http.Error(w, `{"error":"fullname is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.userRepo.UpdateProfile(userID, req); err != nil {
		http.Error(w, `{"error":"failed to update profile"}`, http.StatusInternalServerError)
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch updated profile"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"old_password and new_password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		http.Error(w, `{"error":"incorrect old password"}`, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		http.Error(w, `{"error":"failed to change password"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "password changed successfully"})
}
