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
