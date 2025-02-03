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
