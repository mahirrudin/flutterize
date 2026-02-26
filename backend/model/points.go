package model

import "time"

type PointTransaction struct {
	ID           uint64    `json:"id"`
	SenderID     uint64    `json:"sender_id"`
	ReceiverID   uint64    `json:"receiver_id"`
	SenderName   string    `json:"sender_name"`
	ReceiverName string    `json:"receiver_name"`
	Amount       int64     `json:"amount"`
	Note         string    `json:"note"`
	CreatedAt    time.Time `json:"created_at"`
}

type TransferRequest struct {
	ReceiverIdentifier string `json:"receiver_identifier"`
	Amount             int64  `json:"amount"`
	Note               string `json:"note"`
}

type TransferResponse struct {
	Transaction PointTransaction `json:"transaction"`
	NewBalance  int64            `json:"new_balance"`
}

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}
