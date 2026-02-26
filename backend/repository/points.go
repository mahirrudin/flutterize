package repository

import (
	"database/sql"
	"fmt"

	"github.com/flutterize/backend/model"
)

type PointsRepository struct {
	db *sql.DB
}

func NewPointsRepository(db *sql.DB) *PointsRepository {
	return &PointsRepository{db: db}
}

func (r *PointsRepository) GetBalance(userID uint64) (int64, error) {
	var balance int64
	err := r.db.QueryRow("SELECT points_balance FROM users WHERE id = ?", userID).Scan(&balance)
	return balance, err
}

func (r *PointsRepository) TransferPoints(senderID, receiverID uint64, amount int64, note string) (*model.PointTransaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check sender balance
	var senderBalance int64
	err = tx.QueryRow("SELECT points_balance FROM users WHERE id = ? FOR UPDATE", senderID).Scan(&senderBalance)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender balance: %w", err)
	}

	if senderBalance < amount {
		return nil, fmt.Errorf("insufficient balance: have %d, need %d", senderBalance, amount)
	}

	// Deduct from sender
	_, err = tx.Exec("UPDATE users SET points_balance = points_balance - ? WHERE id = ?", amount, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct from sender: %w", err)
	}

	// Add to receiver
	_, err = tx.Exec("UPDATE users SET points_balance = points_balance + ? WHERE id = ?", amount, receiverID)
	if err != nil {
		return nil, fmt.Errorf("failed to add to receiver: %w", err)
	}

	// Insert transaction record
	result, err := tx.Exec(
		"INSERT INTO point_transactions (sender_id, receiver_id, amount, note) VALUES (?, ?, ?, ?)",
		senderID, receiverID, amount, note,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert transaction: %w", err)
	}

	txID, _ := result.LastInsertId()

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch the created transaction
	transaction := &model.PointTransaction{}
	err = r.db.QueryRow(`
		SELECT pt.id, pt.sender_id, pt.receiver_id, s.fullname, recv.fullname, pt.amount, COALESCE(pt.note, ''), pt.created_at
		FROM point_transactions pt
		JOIN users s ON pt.sender_id = s.id
		JOIN users recv ON pt.receiver_id = recv.id
		WHERE pt.id = ?`, txID).Scan(
		&transaction.ID, &transaction.SenderID, &transaction.ReceiverID,
		&transaction.SenderName, &transaction.ReceiverName,
		&transaction.Amount, &transaction.Note, &transaction.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}

	return transaction, nil
}

func (r *PointsRepository) GetTransactionHistory(userID uint64) ([]model.PointTransaction, error) {
	query := `
		SELECT pt.id, pt.sender_id, pt.receiver_id, s.fullname, recv.fullname, pt.amount, COALESCE(pt.note, ''), pt.created_at
		FROM point_transactions pt
		JOIN users s ON pt.sender_id = s.id
		JOIN users recv ON pt.receiver_id = recv.id
		WHERE pt.sender_id = ? OR pt.receiver_id = ?
		ORDER BY pt.created_at DESC`

	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []model.PointTransaction
	for rows.Next() {
		var t model.PointTransaction
		err := rows.Scan(
			&t.ID, &t.SenderID, &t.ReceiverID,
			&t.SenderName, &t.ReceiverName,
			&t.Amount, &t.Note, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
