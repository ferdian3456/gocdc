package web

type WebResponse struct {
	Code   int         `json:"code"`
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type TokenResponse struct {
	Access_Token  string `json:"access_token"`
	Refresh_Token string `json:"refresh_token"`
}

type ExistenceApiResponse struct {
	Code   int                `json:"code"`
	Status string             `json:"status"`
	Data   ExistenceDataField `json:"data"`
}

type ExistenceDataField struct {
	Status string `json:"status"`
}

type EmailApiResponse struct {
	Code   int            `json:"code"`
	Status string         `json:"status"`
	Data   EmailDataField `json:"data"`
}

type EmailDataField struct {
	Email string `json:"status"`
}
