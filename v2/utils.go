package preloader

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GetJwt(userId uint, jwtKey string, exp time.Duration) (string, error) {
	return jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userId": userId,
			"exp":    time.Now().Add(exp).Unix(),
		},
	).SignedString([]byte(jwtKey))
}
