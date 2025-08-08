package jwt

import "github.com/golang-jwt/jwt/v5"

func NewParser() *jwt.Parser {
	signingMethod := "RS256"
	if jwt.SigningMethodRS256 != nil {
		signingMethod = jwt.SigningMethodRS256.Name
	}
	return jwt.NewParser(jwt.WithValidMethods([]string{signingMethod}))
}
