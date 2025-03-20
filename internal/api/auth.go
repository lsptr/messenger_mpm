package api

import (
	"MessengerMPM/internal/auth"
	"MessengerMPM/internal/db"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Регистрация пользователя

func Register(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Сохраняем пользователя в базе данных
	_, err = db.Pool.Exec(r.Context(), "INSERT INTO users (username, password_hash) VALUES ($1, $2)", user.Username, hashedPassword)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "User registered"}`))
}

// Вход пользователя

func Login(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Проверка пользователя в базе данных
	var userID int
	var passwordHash string
	err := db.Pool.QueryRow(r.Context(), "SELECT id, password_hash FROM users WHERE username = $1", user.Username).Scan(&userID, &passwordHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Генерация токена
	token, err := auth.GenerateToken(userID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Сохранение сессии в базе данных
	_, err = db.Pool.Exec(r.Context(), "INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)",
		userID, token, time.Now().Add(24*time.Hour))
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Возвращаем токен в теле ответа
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"token": "` + token + `", "userID": ` + strconv.Itoa(userID) + `, "username": "` + user.Username + `"}`))

}

// Выход пользователя

func Logout(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из куки
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString := cookie.Value

	// Удаляем сессию из базы данных
	_, err = db.Pool.Exec(r.Context(), "DELETE FROM sessions WHERE token = $1", tokenString)
	if err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	// Удаляем куку token на стороне клиента
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour), // Устанавливаем срок действия в прошлое
		Path:    "/",
	})

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged out"}`))
}
