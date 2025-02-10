package product

import "time"

type ProductEvent struct {
	Name       string     `json:"name"`
	Price      float64    `json:"price"`
	Seller_id  string     `json:"seller_id"`
	Created_at *time.Time `json:"created_at"`
	Updated_at *time.Time `json:"updated_at"`
}
