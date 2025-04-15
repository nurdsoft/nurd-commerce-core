package entities

type Oauth2Response struct {
	AccessToken      string `json:"access_token"`
	InstanceURL      string `json:"instance_url"`
	ID               string `json:"id"`
	TokenType        string `json:"token_type"`
	IssuedAt         string `json:"issued_at"`
	Signature        string `json:"signature"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
