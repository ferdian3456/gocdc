package product

type ProductUpdateRequest struct {
	Name        string  `validate:"omitempty,min=5,max=20" json:"name,omitempty"`
	Quantity    int     `validate:"omitempty,min=1" json:"quantity,omitempty"`
	Price       float64 `validate:"omitempty,min=1" json:"price,omitempty"`
	Weight      int     `validate:"omitempty,min=1" json:"weight,omitempty"`
	Size        string  `validate:"omitempty,min=1,max=4" json:"size,omitempty"`
	Status      string  `validate:"omitempty,min=5,max=9" json:"status,omitempty"`
	Description string  `validate:"omitempty,min=10" json:"description,omitempty"`
}
