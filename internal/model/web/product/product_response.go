package product

import "time"

type ProductResponse struct {
	Id          int        `json:"id"`
	Seller_id   string     `json:"seller_id"`
	Name        string     `json:"name"`
	Quantity    int        `json:"quantity"`
	Price       float64    `json:"price"`
	Weight      int        `json:"weight"`
	Size        string     `json:"size"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	Created_at  *time.Time `json:"created_at"`
	Updated_at  *time.Time `json:"updated_at"`
}

type ProductHomePageResponse struct {
	Id             int        `json:"id"`
	Seller_id      string     `json:"seller_id"`
	Seller_name    string     `json:"seller_name"`
	Seller_address string     `json:"seller_address"`
	Name           string     `json:"name"`
	Quantity       int        `json:"quantity"`
	Price          float64    `json:"price"`
	Weight         int        `json:"weight"`
	Size           string     `json:"size"`
	Status         string     `json:"status"`
	Description    string     `json:"description"`
	Created_at     *time.Time `json:"created_at"`
	Updated_at     *time.Time `json:"updated_at"`
}
