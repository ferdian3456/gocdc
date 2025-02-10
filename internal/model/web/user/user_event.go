package user

import "time"

type UserEvents struct {
	Id         string     `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Created_at *time.Time `json:"created_at"`
	Updated_at *time.Time `json:"updated_at"`
}
