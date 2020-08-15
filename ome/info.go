package ome

type Endpoints struct {
	Authorize string `json:"authorize"`
	Token     string `json:"token"`
	Revoke    string `json:"revoke"`
	Verify    string `json:"verify"`
}

type DataService struct {
	GRPC string `json:"grpc"`
	HTTP string `json:"http"`
}

type Oauth2Config struct {
	Endpoints    Endpoints `json:"endpoints"`
	SignatureKey string    `json:"signature_key"`
}

type Info struct {
	Registration string       `json:"registration"`
	CSR          string       `json:"csr"`
	Data         DataService  `json:"data"`
	Oauth2       Oauth2Config `json:"oauth2"`
}
