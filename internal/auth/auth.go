package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("messenger_server_key")

// создает JWT токен для пользователя.
func GenerateToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(), // Токен действителен 24 часа
	})
	return token.SignedString(jwtKey)
}

// проверяет JWT токен и возвращает ID пользователя.
func ValidateToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return 0, err
	}
	claims := token.Claims.(jwt.MapClaims)
	return int(claims["user_id"].(float64)), nil
}
