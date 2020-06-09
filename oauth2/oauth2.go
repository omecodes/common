package oauth2

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/omecodes/common/crypto"
	"github.com/omecodes/common/jcon"
	"github.com/omecodes/common/log"
	"github.com/omecodes/common/system"
	"net/http"
	"net/url"
	"strings"
	"syscall"
)

const (
	ParamClientID          = "client_id"
	ParamNonce             = "nonce"
	ParamClientAuthMessage = "auth_message"
	ParamResponseType      = "response_type"
	ParamState             = "state"
	ParamScope             = "scope"
	ParamRedirectURI       = "redirect_uri"
	ParamCode              = "code"
	ParamProvider          = "provider"
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

	KeyAccessToken  = "access_token"
	KeyTokenType    = "token_type"
	KeyExpiresIn    = "expires_in"
	KeyIdToken      = "id_token"
	KeyRefreshToken = "refresh_token"
	KeyScope        = "scope"
)

type Provider interface {
	AuthorizeURL() string
	AccessTokenURL() string
	EncodedCertificate() string
	SignatureAlg() string
	JWK() string
}

type Config struct {
	ClientID              string
	Secret                string
	Scope                 []string
	State                 string
	CallbackURL           string
	AuthorizationEndpoint string
	TokenEndpoint         string
}

// Client
type Client struct {
	cfg *Config
}

func NewClient(cfg *Config) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) GetURLAuthorizationURL() (string, error) {
	authURL := fmt.Sprintf("%s?%s=%s&%s=%s&%s=%s&%s=%s",
		c.cfg.AuthorizationEndpoint,
		ParamState,
		c.cfg.State,
		ParamScope,
		strings.Join(c.cfg.Scope, " "),
		ParamClientID,
		c.cfg.ClientID,
		ParamResponseType,
		ResponseTypeCode)

	if c.cfg.CallbackURL != "" {
		authURL = fmt.Sprintf("%s&%s=%s", authURL, ParamRedirectURI, url.QueryEscape(c.cfg.CallbackURL))
	}
	return authURL, nil
}

func (c *Client) GetAccessToken(code string) (*Token, error) {
	client := &http.Client{}

	form := url.Values{}
	form.Add(ParamCode, code)
	form.Add(ParamGrantType, GrantTypeAuthorizationCode)

	req, err := http.NewRequest(http.MethodPost, c.cfg.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	auth := c.cfg.ClientID + ":" + c.cfg.Secret
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if rsp.StatusCode == 200 {
		var token Token
		err = json.NewDecoder(rsp.Body).Decode(&token)
		return &token, err
	} else {
		m := jcon.Map{}
		err = json.NewDecoder(rsp.Body).Decode(&m)
		return nil, errors.New(m.String())
	}
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
	code        chan string
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

	code := query.Get(ParamCode)
	if code != "" {
		h.code <- code

	} else {
		err := query.Get(ParamError)
		errDescription := query.Get(ParamErrorDescription)
		h.errorChan <- errors.New(err + ". details: " + errDescription)
	}
}

func (h *RedirectURIHandler) GetCode() (string, error) {
	go h.listen()

	select {
	case e := <-h.errorChan:
		return "", e

	case jwt := <-h.code:
		return jwt, nil
	}
}

func NewRedirectURIHandler(redirectURI string, tc *tls.Config) *RedirectURIHandler {
	return &RedirectURIHandler{
		redirectURI: redirectURI,
		tlsConfigs:  tc,
		errorChan:   make(chan error, 1),
		code:        make(chan string, 1),
	}
}

func GetToken(cfg *Config) (*Token, error) {
	client := NewClient(cfg)
	authorizationURL, err := client.GetURLAuthorizationURL()
	if err != nil {
		log.Error("could not construct authorization URL", err)
	}

	cmd := system.OpenBrowserCMD(authorizationURL)
	if cmd == nil {
		return nil, errors.New("could not open browser")
	}

	go func() {
		err := cmd.Start()
		if err != nil {
			log.Error("run browser command failed", err)
		}
	}()

	redHandler := NewRedirectURIHandler("http://localhost:9876/handle", nil)
	code, err := redHandler.GetCode()
	if err != nil {
		return nil, err
	}

	_ = cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Process.Kill()
	return client.GetAccessToken(code)
}
