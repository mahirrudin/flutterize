package repository

import (
	"database/sql"
	"time"

	"github.com/flutterize/backend/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmailOrPhone(identifier string) (*model.User, error) {
	query := `SELECT id, email, phone, fullname, birthdate, password, points_balance,
		reset_token, reset_token_expiry, created_at, updated_at
		FROM users WHERE email = ? OR phone = ?`

	user := &model.User{}
	var birthdate string
	err := r.db.QueryRow(query, identifier, identifier).Scan(
		&user.ID, &user.Email, &user.Phone, &user.Fullname, &birthdate,
		&user.Password, &user.PointsBalance, &user.ResetToken, &user.ResetTokenExpiry,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Birthdate = birthdate
	return user, nil
}

func (r *UserRepository) FindByID(id uint64) (*model.User, error) {
	query := `SELECT id, email, phone, fullname, birthdate, password, points_balance,
		reset_token, reset_token_expiry, created_at, updated_at
		FROM users WHERE id = ?`

	user := &model.User{}
	var birthdate string
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Phone, &user.Fullname, &birthdate,
		&user.Password, &user.PointsBalance, &user.ResetToken, &user.ResetTokenExpiry,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Birthdate = birthdate
	return user, nil
}

func (r *UserRepository) UpdateProfile(id uint64, req model.UpdateProfileRequest) error {
	query := `UPDATE users SET fullname = ?, phone = ?, birthdate = ? WHERE id = ?`
	_, err := r.db.Exec(query, req.Fullname, req.Phone, req.Birthdate, id)
	return err
}

func (r *UserRepository) UpdatePassword(id uint64, hashedPassword string) error {
	query := `UPDATE users SET password = ? WHERE id = ?`
	_, err := r.db.Exec(query, hashedPassword, id)
	return err
}

func (r *UserRepository) SetResetToken(id uint64, token string, expiry time.Time) error {
	query := `UPDATE users SET reset_token = ?, reset_token_expiry = ? WHERE id = ?`
	_, err := r.db.Exec(query, token, expiry, id)
	return err
}

func (r *UserRepository) FindByResetToken(token string) (*model.User, error) {
	query := `SELECT id, email, phone, fullname, birthdate, password, points_balance,
		reset_token, reset_token_expiry, created_at, updated_at
		FROM users WHERE reset_token = ? AND reset_token_expiry > NOW()`

	user := &model.User{}
	var birthdate string
	err := r.db.QueryRow(query, token).Scan(
		&user.ID, &user.Email, &user.Phone, &user.Fullname, &birthdate,
		&user.Password, &user.PointsBalance, &user.ResetToken, &user.ResetTokenExpiry,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	user.Birthdate = birthdate
	return user, nil
}

func (r *UserRepository) Create(req model.RegisterRequest, hashedPassword string) (*model.User, error) {
	result, err := r.db.Exec(
		`INSERT INTO users (email, phone, fullname, birthdate, password, points_balance) VALUES (?, ?, ?, ?, ?, 0)`,
		req.Email, req.Phone, req.Fullname, req.Birthdate, hashedPassword,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return r.FindByID(uint64(id))
}

func (r *UserRepository) ClearResetToken(id uint64) error {
	query := `UPDATE users SET reset_token = NULL, reset_token_expiry = NULL WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
