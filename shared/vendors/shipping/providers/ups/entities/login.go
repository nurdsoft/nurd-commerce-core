package entities

type Oauth2Response struct {
	AccessToken string `json:"access_token"`
	ClientID    string `json:"client_id"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
	ExpiresIn   string `json:"expires_in"`
	Status      string `json:"status"`
}
