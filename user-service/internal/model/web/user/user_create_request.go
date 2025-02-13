package user

type UserRegisterRequest struct {
	Name            string `validate:"required,min=5,max=20" json:"name"`
	Profile_picture string `validate:"required,min=20,max=255" json:"profile_picture"`
	Email           string `validate:"required,min=5,max=254" json:"email"`
	Password        string `validate:"required,min=5,max=20" json:"password"`
	Address         string `validate:"required,min=10,max=30" json:"address"`
	PhoneNumber     string `validate:"required,min=12,max=12" json:"phone_number"`
}

type RenewalTokenRequest struct {
	Refresh_token string `validate:"required,min=43" json:"refresh_token"`
}
