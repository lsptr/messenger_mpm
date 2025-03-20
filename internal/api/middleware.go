package api

import (
	"MessengerMPM/internal/db"
	"MessengerMPM/internal/rabbitmq"
	"MessengerMPM/internal/websocket"
	"context"
	"net/http"
)

// Проверка JWT токена и добавление ID пользователя в контекст

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Убедитесь, что токен начинается с "Bearer "
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString = tokenString[7:] // Убираем "Bearer "

		// Проверка токена в базе данных
		var userID int
		err := db.Pool.QueryRow(r.Context(), "SELECT user_id FROM sessions WHERE token = $1 AND expires_at > NOW()", tokenString).Scan(&userID)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Добавление ID пользователя в контекст
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Проверяем, что текущий пользователь является администратором чата

func IsAdmin(chatID string, userID int) (bool, error) {
	var adminID int
	err := db.Pool.QueryRow(context.Background(), `
        SELECT admin_id FROM chats WHERE id = $1`, chatID).Scan(&adminID)
	if err != nil {
		return false, err
	}
	return adminID == userID, nil
}

// Получаем ID пользователя по его имени

func GetUser(username string) (int, error) {
	var userId int
	err := db.Pool.QueryRow(context.Background(), `
        SELECT id FROM users WHERE username = $1`, username).Scan(&userId)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

var (
	hub    *websocket.Hub
	rabbit *rabbitmq.RabbitMQ
)

// InitAPI инициализирует API с зависимостями.
func InitAPI(h *websocket.Hub, r *rabbitmq.RabbitMQ) {
	hub = h
	rabbit = r
}
