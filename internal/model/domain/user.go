package domain

import "time"

type User struct {
	Id          string
	Name        string
	Email       string
	Password    string
	Address     string
	PhoneNumber string
	Created_at  *time.Time
	Updated_at  *time.Time
}

type RefreshToken struct {
	User_id              string
	Hashed_refresh_token string
	Created_at           *time.Time
	Expired_at           *time.Time
}
