package authpb

import (
	"github.com/gorilla/securecookie"
)

const (
	jwtKey = "jwt"
)

func ExtractJwtFromAccessToken(name, value string, codecs ...securecookie.Codec) (string, error) {
	values := make(map[interface{}]interface{})
	err := securecookie.DecodeMulti(name, value, &values, codecs...)
	if err != nil {
		return "", err
	}
	o := values[jwtKey]
	if o == nil {
		return "", nil
	}
	return o.(string), nil
}

func JwtEmbeddedAccessToken(name, jwt string, codecs ...securecookie.Codec) (string, error) {
	values := make(map[interface{}]interface{})
	values[jwtKey] = jwt
	return securecookie.EncodeMulti(name, values, codecs...)
}

type CookieJWTBearer struct{}
