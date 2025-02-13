package user

import "time"

type UserResponse struct {
	Id          string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Address     string     `json:"address"`
	PhoneNumber string     `json:"phone_number"`
	Created_at  *time.Time `json:"created_at"`
	Updated_at  *time.Time `json:"updated_at"`
}
