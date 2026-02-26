package model

import "time"

type User struct {
	ID               uint64     `json:"id"`
	Email            string     `json:"email"`
	Phone            string     `json:"phone"`
	Fullname         string     `json:"fullname"`
	Birthdate        string     `json:"birthdate"`
	Password         string     `json:"-"`
	PointsBalance    int64      `json:"points_balance"`
	ResetToken       *string    `json:"-"`
	ResetTokenExpiry *time.Time `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateProfileRequest struct {
	Fullname  string `json:"fullname"`
	Phone     string `json:"phone"`
	Birthdate string `json:"birthdate"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type ForgotPasswordRequest struct {
	Identifier string `json:"identifier"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RegisterRequest struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Fullname      string `json:"fullname"`
	Birthdate     string `json:"birthdate"`
	Password      string `json:"password"`
	PointsBalance int64  `json:"points_balance"`
}
