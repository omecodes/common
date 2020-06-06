package oauth2

type Token struct {
	IdToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   uint32 `json:"expires_in"`
	Scope       string `json:"scope"`
}
