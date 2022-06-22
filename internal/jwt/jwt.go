package jwt

import (
	"fmt"
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

func GetUserID(tokenString string, key string) (int, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return int(claims["id"].(float64)), nil
	} else {
		return -1, err
	}
}
