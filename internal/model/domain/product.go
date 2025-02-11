package domain

import "time"

type Product struct {
	Id              int
	Seller_id       string
	Name            string
	Product_picture string
	Quantity        int
	Price           float64
	Weight          int
	Size            string
	Status          string
	Description     string
	Created_at      *time.Time
	Updated_at      *time.Time
}
