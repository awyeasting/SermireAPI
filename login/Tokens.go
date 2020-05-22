package login

import (
	jwt "github.com/dgrijalva/jwt-go"

	"fmt"
)

func GetClaims(tokenString string) (jwt.MapClaims, error){
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate alg is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(JWT_SIGNING_SECRET), nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		return claims, nil
	}
	return nil, err
}