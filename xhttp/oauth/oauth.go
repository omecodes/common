package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/zoenion/common/crypto"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Provider interface {
	AuthorizeURL() string
	AccessTokenURL() string
	EncodedCertificate() string
	SignatureAlg() string
	JWK() string
}

const (
	ParamClientID          = "client_id"
	ParamNonce             = "nonce"
	ParamClientAuthMessage = "auth_message"
	ParamState             = "state"
	ParamScope             = "scope"
	ParamRedirectURI       = "redirect_uri"
	ParamCode              = "code"
)

type urlBuilder struct {
	state        string
	scope        string
	redirectURI  string
	clientID     string
	clientSecret string
}

func (rb *urlBuilder) setAccess(clientID, secret string) *urlBuilder {
	rb.clientID = clientID
	rb.clientSecret = secret
	return rb
}

func (rb *urlBuilder) SetState(state string) *urlBuilder {
	rb.state = state
	return rb
}

func (rb *urlBuilder) SetScope(scope string) *urlBuilder {
	rb.scope = scope
	return rb
}

func (rb *urlBuilder) SetRedirectURI(uri string) *urlBuilder {
	rb.redirectURI = uri
	return rb
}

func (rb *urlBuilder) URL(providerURL string) (string, error) {

	authMessage, nonce, err := CreateAuth(rb.clientSecret)
	if err != nil {
		return "", err
	}

	encodedRedirectURI := base64.URLEncoding.EncodeToString([]byte(rb.redirectURI))
	return fmt.Sprintf("%s?%s=%s&%s=%s&%s=%s&%s=%s&%s=%s&%s=%s",
		providerURL,
		ParamClientID,
		rb.clientID,
		ParamState,
		rb.state,
		ParamScope,
		rb.scope,
		ParamNonce,
		nonce,
		ParamClientAuthMessage,
		authMessage,
		ParamRedirectURI,
		encodedRedirectURI,
	), nil
}

// Client
type Client struct {
	state    string
	provider Provider
}

func (c *Client) GetURL(clientID, clientSecret, scope, redirectURI string) (string, error) {
	state := make([]byte, 16)
	_, err := rand.Read(state)
	if err != nil {
		return "", err
	}
	c.state = string(state)
	return new(urlBuilder).setAccess(clientID, clientSecret).SetRedirectURI(redirectURI).SetScope(scope).SetState(c.state).URL(c.provider.AuthorizeURL())
}

func (c *Client) GetAccessToken(challenge string, clientID, clientSecret string) (string, error) {
	key, _ := base64.StdEncoding.DecodeString(clientSecret)

	challengeBytes, err := base64.URLEncoding.DecodeString(challenge)
	if err != nil {
		return "", err
	}

	decrypted, err := crypto2.AESGCMDecrypt(key, challengeBytes)
	if err != nil {
		return "", err
	}

	salt := decrypted[:12]
	code := decrypted[12:]
	encryptedCode, err := crypto2.AESGCMEncryptWithSalt(key, salt, code)
	if err != nil {
		return "", err
	}

	encodedCode := base64.URLEncoding.EncodeToString(encryptedCode)
	url := fmt.Sprintf("%s?%s=%s&%s=%s", c.provider.AccessTokenURL(), ParamCode, encodedCode, ParamClientID, clientID)

	rsp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	if rsp.StatusCode == 200 {
		bodyBytes, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return "", err
		}
		return string(bodyBytes), nil
	}

	return "", errors.New("failed to get jwt")
}

func (c *Client) GetState() string {
	return c.state
}

// NewClient
func NewClient(provider Provider) *Client {
	c := new(Client)
	c.provider = provider
	return c
}

// Params
type AuthorizeParams struct {
	ClientID    string
	AuthMessage string
	Nonce       string
	State       string
	Scope       string
	RedirectURI string
}

func (p *AuthorizeParams) FromURL(u *url.URL) error {
	query := u.Query()

	p.ClientID = query.Get(ParamClientID)
	if p.ClientID == "" {
		return errors.New("missing " + ParamClientID)
	}

	p.AuthMessage = query.Get(ParamClientAuthMessage)
	if p.ClientID == "" {
		return errors.New("missing " + ParamClientAuthMessage)
	}

	p.Scope = query.Get(ParamScope)
	if p.Scope == "" {
		return errors.New("missing " + ParamScope)
	}

	p.Nonce = query.Get(ParamNonce)
	if p.ClientID == "" {
		return errors.New("missing " + ParamNonce)
	}

	p.State = query.Get(ParamState)
	if p.ClientID == "" {
		return errors.New("missing " + ParamState)
	}

	p.RedirectURI = query.Get(ParamRedirectURI)
	if p.ClientID == "" {
		return errors.New("missing " + ParamRedirectURI)
	}

	return nil
}

type GetAccessTokenParams struct {
	ClientID string
	Code     string
}

func (p *GetAccessTokenParams) FromURL(u *url.URL) error {
	return nil
}

func CreateAuth(secret string) (string, string, error) {
	nonceBytes := make([]byte, 16)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", "", err
	}

	nonce := base64.URLEncoding.EncodeToString(nonceBytes)
	secretBytes := []byte(secret)
	m := hmac.New(sha512.New, secretBytes)
	m.Write(nonceBytes)
	mm := m.Sum(nil)
	return base64.StdEncoding.EncodeToString(mm), nonce, nil
}

func VerifyAuth(secret, nonce, authMessage string) (bool, error) {
	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return false, err
	}

	m := hmac.New(sha512.New, []byte(secret))
	m.Write(nonceBytes)
	mm := m.Sum(nil)

	return string(mm) == authMessage, nil
}
