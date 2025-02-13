package user

type UserUpdateRequest struct {
	Name        string `validate:"omitempty,min=5,max=20" json:"name,omitempty"`
	Email       string `validate:"omitempty,min=5,max=254" json:"email,omitempty"`
	Password    string `validate:"omitempty,min=5,max=20" json:"password,omitempty"`
	Address     string `validate:"omitempty,min=10,max=30" json:"address,omitempty"`
	PhoneNumber string `validate:"omitempty,min=12,max=12" json:"phone_number,omitempty"`
}
