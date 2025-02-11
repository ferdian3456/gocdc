package product

import "time"

type ProductEvent struct {
	Id         string     `json:"id"`
	Email      string     `json:"email"`
	Event      string     `json:"event"`
	Created_at *time.Time `json:"created_at"`
}
