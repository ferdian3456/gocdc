package product

type ProductCreateRequest struct {
	Name            string  `validate:"required,min=5,max=20" json:"name"`
	Product_picture string  `validate:"required,min=20,max=255" json:"product_picture"`
	Quantity        int     `validate:"required,min=1" json:"quantity"`
	Price           float64 `validate:"required,min=1" json:"price"`
	Weight          int     `validate:"required,min=1" json:"weight"`
	Size            string  `validate:"required,min=1,max=4" json:"size"`
	Description     string  `validate:"required,min=10" json:"description"`
}
