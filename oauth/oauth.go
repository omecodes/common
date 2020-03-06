package oauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/zoenion/common/crypto"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	ParamResponseType      = "response_type"
	ParamState             = "state"
	ParamScope             = "scope"
	ParamRedirectURI       = "redirect_uri"
	ParamCode              = "code"
	ParamAlg               = "algorithm"
	ParamError             = "error"
	ParamErrorDescription  = "error_description"
	ParamErrorUri          = "error_uri"
	ParamGrantType         = "grant_type"
	ParamCodeVerifier      = "code_verifier"

	ResponseTypeCode         = "code"
	ResponseTypeIDToken      = "id_token"
	ResponseTypeTokenIDToken = "token id_token"

	ErrorInvalidRequest          = "invalid_request"
	ErrorUnauthorizedClient      = "unauthorized_client"
	ErrorAccessDenied            = "access_denied"
	ErrorUnsupportedResponseType = "unsupported_response_type"
	ErrorInvalidScope            = "invalid_scope"
	ErrorServerError             = "server_error"
	ErrorTemporarilyUnavailable  = "temporarily_unavailable"

	GrantTypeAuthorizationCode = "authorization_code"
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

	params := url.Values{}
	params.Add(ParamClientID, rb.clientID)
	params.Add(ParamState, rb.state)
	params.Add(ParamScope, rb.scope)
	params.Add(ParamNonce, nonce)
	params.Add(ParamClientAuthMessage, authMessage)
	params.Add(ParamRedirectURI, rb.redirectURI)
	return fmt.Sprintf("%s?%s", providerURL, params.Encode()), nil
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
	c.state = hex.EncodeToString(state)
	return new(urlBuilder).setAccess(clientID, clientSecret).SetRedirectURI(redirectURI).SetScope(scope).SetState(c.state).URL(c.provider.AuthorizeURL())
}

func (c *Client) GetAccessToken(challenge string, clientID, clientSecret string) (string, error) {
	key, err := hex.DecodeString(clientSecret)
	if err != nil {
		return "", err
	}

	challengeBytes, err := hex.DecodeString(challenge)
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
	encodedCode := hex.EncodeToString(encryptedCode)

	params := url.Values{}
	params.Add(ParamCode, encodedCode)
	params.Add(ParamClientID, clientID)
	u := fmt.Sprintf("%s?%s", c.provider.AccessTokenURL(), params.Encode())

	rsp, err := http.Get(u)
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
	ClientID     string
	ResponseType string
	State        string
	Scope        string
	RedirectURI  string
	Missing      []string
}

func (p *AuthorizeParams) Has(param string) bool {
	for _, p := range p.Missing {
		if p == param {
			return false
		}
	}
	return true
}

func (p *AuthorizeParams) FromURL(u *url.URL) bool {
	query := u.Query()

	uStr := u.String()
	if !strings.Contains(uStr, ParamClientID) &&
		!strings.Contains(uStr, ParamResponseType) &&
		!strings.Contains(uStr, ParamScope) &&
		!strings.Contains(uStr, ParamState) &&
		!strings.Contains(uStr, ParamRedirectURI) {
		return false
	}

	p.ClientID = query.Get(ParamClientID)
	if p.ClientID == "" {
		p.Missing = append(p.Missing, ParamClientID)
	}

	p.ResponseType = query.Get(ParamResponseType)
	if p.ResponseType == "" {
		p.Missing = append(p.Missing, ParamResponseType)
	}

	p.Scope = query.Get(ParamScope)
	if p.Scope == "" {
		p.Missing = append(p.Missing, ParamScope)
	}

	p.State = query.Get(ParamState)
	if p.State == "" {
		p.Missing = append(p.Missing, ParamState)
	}

	p.RedirectURI = query.Get(ParamRedirectURI)
	if p.RedirectURI == "" {
		p.Missing = append(p.Missing, ParamRedirectURI)
	}
	return true
}

func (p *AuthorizeParams) ToMap() map[string]string {
	return map[string]string{
		ParamState:        p.State,
		ParamClientID:     p.ClientID,
		ParamScope:        p.Scope,
		ParamResponseType: p.ResponseType,
		ParamRedirectURI:  p.RedirectURI,
	}
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

	nonce := hex.EncodeToString(nonceBytes)
	secretBytes := []byte(secret)
	m := hmac.New(sha512.New, secretBytes)
	m.Write(nonceBytes)
	mm := m.Sum(nil)
	return hex.EncodeToString(mm), nonce, nil
}

func VerifyAuth(secret, nonce, authMessage string) (bool, error) {
	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil {
		return false, err
	}

	m := hmac.New(sha512.New, []byte(secret))
	m.Write(nonceBytes)
	mm := m.Sum(nil)

	calculatedAuth := hex.EncodeToString(mm)
	result := calculatedAuth == authMessage
	return result, nil
}

// CodeChallenge
type CodeChallenge struct {
	Alg           string
	EncryptedCode string
}

func (c *CodeChallenge) ProcessChallenge(secret string) ([]byte, error) {
	if c.Alg == "aes-gcm-256" {
		keyBytes, err := hex.DecodeString(secret)
		if err != nil {
			return nil, err
		}

		challengeData, err := hex.DecodeString(c.EncryptedCode)
		if err != nil {
			return nil, err
		}

		codeDataBytes, err := crypto2.AESGCMDecrypt(keyBytes, challengeData)
		if err != nil {
			return nil, err
		}

		salt := codeDataBytes[:12]
		codeBytes := codeDataBytes[12:]

		return crypto2.AESGCMEncryptWithSalt(keyBytes, salt, codeBytes)
	}
	return nil, errors.New("unsupported algorithm")
}

func CreateCodeChallenge(secret string) (*CodeChallenge, string, error) {
	codeBytes := make([]byte, 16)
	_, err := rand.Read(codeBytes)
	if err != nil {
		return nil, "", err
	}

	key, err := hex.DecodeString(secret)
	if err != nil {
		return nil, "", err
	}

	clientSalt := make([]byte, 12)
	salt := make([]byte, 12)

	_, err = rand.Read(salt)
	if err != nil {
		return nil, "", err
	}
	_, err = rand.Read(clientSalt)
	if err != nil {
		return nil, "", err
	}

	encryptedCode, err := crypto2.AESGCMEncryptWithSalt(key, salt, append(clientSalt, codeBytes...))
	if err != nil {
		return nil, "", err
	}
	codeChallenge := hex.EncodeToString(encryptedCode)

	return &CodeChallenge{
		Alg:           "aes-gcm-256",
		EncryptedCode: codeChallenge,
	}, hex.EncodeToString(codeBytes), nil
}

// CodeChallengeResult
type CodeChallengeResult struct {
	Alg           string
	ClientID      string
	EncryptedCode string
}

func (c *CodeChallengeResult) GetCode(secret string) ([]byte, error) {
	if c.Alg == "aes-gcm-256" {
		keyBytes, err := hex.DecodeString(secret)
		if err != nil {
			return nil, err
		}

		encryptedCodeData, err := hex.DecodeString(c.EncryptedCode)
		if err != nil {
			return nil, err
		}
		return crypto2.AESGCMDecrypt(keyBytes, encryptedCodeData)
	}
	return nil, errors.New("unsupported algorithm")
}

func (c *CodeChallengeResult) FromURL(u *url.URL) error {
	c.EncryptedCode = u.Query().Get(ParamCode)
	if c.EncryptedCode == "" {
		return errors.New("missing " + ParamCode)
	}

	c.ClientID = u.Query().Get(ParamClientID)
	if c.ClientID == "" {
		return errors.New("missing " + ParamClientID)
	}

	c.ClientID = u.Query().Get(ParamAlg)
	return nil
}

// RedirectURIHandler
type RedirectURIHandler struct {
	redirectURI string
	tlsConfigs  *tls.Config
	errorChan   chan error
	jwtChan     chan string
}

func (h *RedirectURIHandler) listen() {
	u, err := url.Parse(h.redirectURI)
	if err != nil {
		h.errorChan <- err
		return
	}

	http.HandleFunc(u.Path, h.handle)
	err = http.ListenAndServe(u.Host, nil)
	if err != nil {
		h.errorChan <- err
	}
}

func (h *RedirectURIHandler) handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	if len(query) == 0 {
		h.errorChan <- errors.New("no token provided")
	}
}

func (h *RedirectURIHandler) GetToken() (string, error) {
	go h.listen()

	select {
	case e := <-h.errorChan:
		return "", e
	case jwt := <-h.jwtChan:
		return jwt, nil
	}
}

func NewRedirectURIHandler(redirectURI string, tc *tls.Config) *RedirectURIHandler {
	return &RedirectURIHandler{
		redirectURI: redirectURI,
		tlsConfigs:  tc,
		errorChan:   make(chan error, 1),
		jwtChan:     make(chan string, 1),
	}
}
