package jwt

import (
	"time"

	"github.com/golang-jwt/jwt"
)

func Generate(userid int, key string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userid,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24).Unix(),
		"nbf": now.Add(time.Second * 4).Unix(),
	})
	tokenString, err := token.SignedString([]byte(key))
	return tokenString, err
}
