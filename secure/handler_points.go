package handler

import (
	"encoding/json"
	"net/http"

	"github.com/flutterize/backend/middleware"
	"github.com/flutterize/backend/model"
	"github.com/flutterize/backend/repository"
)

type PointsHandler struct {
	pointsRepo *repository.PointsRepository
	userRepo   *repository.UserRepository
}

func NewPointsHandler(pointsRepo *repository.PointsRepository, userRepo *repository.UserRepository) *PointsHandler {
	return &PointsHandler{pointsRepo: pointsRepo, userRepo: userRepo}
}

func (h *PointsHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	balance, err := h.pointsRepo.GetBalance(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to get balance"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.BalanceResponse{Balance: balance})
}

func (h *PointsHandler) TransferPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	senderID := middleware.GetUserID(r)

	var req model.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.ReceiverIdentifier == "" || req.Amount <= 0 {
		http.Error(w, `{"error":"receiver_identifier and positive amount are required"}`, http.StatusBadRequest)
		return
	}

	// Find receiver by email or phone
	receiver, err := h.userRepo.FindByEmailOrPhone(req.ReceiverIdentifier)
	if err != nil {
		http.Error(w, `{"error":"receiver not found"}`, http.StatusBadRequest)
		return
	}

	if receiver.ID == senderID {
		http.Error(w, `{"error":"cannot transfer to yourself"}`, http.StatusBadRequest)
		return
	}

	transaction, err := h.pointsRepo.TransferPoints(senderID, receiver.ID, req.Amount, req.Note)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	newBalance, _ := h.pointsRepo.GetBalance(senderID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.TransferResponse{
		Transaction: *transaction,
		NewBalance:  newBalance,
	})
}

func (h *PointsHandler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	transactions, err := h.pointsRepo.GetTransactionHistory(userID)
	if err != nil {
		http.Error(w, `{"error":"failed to get transaction history"}`, http.StatusInternalServerError)
		return
	}

	if transactions == nil {
		transactions = []model.PointTransaction{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
