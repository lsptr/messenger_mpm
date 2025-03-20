package api

import (
	"MessengerMPM/internal/db"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

//Создание чата

func CreateChat(w http.ResponseWriter, r *http.Request) {
	var chat struct {
		Name      string `json:"name"`
		Algorithm string `json:"algorithm"`
		Username  string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&chat); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if chat.Name == "" || chat.Algorithm == "" || chat.Username == "" {
		http.Error(w, "Name, algorithm, and username are required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	user2ID, err := GetUser(chat.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find user", http.StatusInternalServerError)
		}
		return
	}

	if userID == user2ID {
		http.Error(w, "You cannot create a chat with yourself", http.StatusBadRequest)
		return
	}

	tx, err := db.Pool.Begin(r.Context())
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())
	var chatID int
	err = tx.QueryRow(r.Context(), `
        INSERT INTO chats (name, algorithm, user1_id, user2_id) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id`, chat.Name, chat.Algorithm, userID, user2ID).Scan(&chatID)
	if err != nil {
		http.Error(w, "Failed to create chat", http.StatusInternalServerError)
		return
	}
	log.Printf("запись о чате добавлена")
	// Создаем очереди для каждого пользователя
	queueName1 := "chat_" + strconv.Itoa(chatID) + "_user_" + strconv.Itoa(userID)
	queueName2 := "chat_" + strconv.Itoa(chatID) + "_user_" + strconv.Itoa(user2ID)

	_, err = rabbit.DeclareQueue(queueName1)
	if err != nil {
		log.Printf("Ошибка при создании очереди %s: %v", queueName1, err)
		http.Error(w, "Failed to create RabbitMQ queue", http.StatusInternalServerError)
		return
	}

	_, err = rabbit.DeclareQueue(queueName2)
	if err != nil {
		log.Printf("Ошибка при создании очереди %s: %v", queueName2, err)
		http.Error(w, "Failed to create RabbitMQ queue", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"chat_id": ` + strconv.Itoa(chatID) + `}`))
}

// Удаление чата

func DeleteChat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	var user1ID, user2ID int
	err := db.Pool.QueryRow(r.Context(), `
        SELECT user1_id, user2_id FROM chats WHERE id = $1`, chatID).Scan(&user1ID, &user2ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Chat not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to check chat", http.StatusInternalServerError)
		}
		return
	}

	if userID != user1ID && userID != user2ID {
		http.Error(w, "You are not a participant of this chat", http.StatusForbidden)
		return
	}

	// Удаляем чат
	_, err = db.Pool.Exec(r.Context(), "DELETE FROM chats WHERE id = $1", chatID)
	if err != nil {
		http.Error(w, "Failed to delete chat", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Chat ` + chatID + ` deleted"}`))
}

// GetUserChats возвращает все чаты, в которых участвует пользователь с указанным user_id.

func GetUserChats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	// Выполняем SQL-запрос для получения чатов
	rows, err := db.Pool.Query(r.Context(), `
        SELECT c.id, c.name, c.algorithm, c.user1_id, c.user2_id, c.created_at, u.username AS user2_name
        FROM chats c
        JOIN users u ON c.user2_id = u.id
        WHERE c.user1_id = $1
        UNION
        SELECT c.id, c.name, c.algorithm, c.user1_id, c.user2_id, c.created_at, u.username AS user2_name
        FROM chats c
        JOIN users u ON c.user1_id = u.id
        WHERE c.user2_id = $1`, userID)
	if err != nil {
		http.Error(w, "Failed to fetch chats", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Создаем структуру для хранения информации о чате
	type Chat struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		Algorithm string    `json:"algorithm"`
		User1ID   int       `json:"user1_id"`
		User2ID   int       `json:"user2_id"`
		CreatedAt time.Time `json:"created_at"`
		User2Name string    `json:"user2_name"`
	}

	var chats []Chat

	// Итерируем по результатам запроса
	for rows.Next() {
		var chat Chat
		if err := rows.Scan(&chat.ID, &chat.Name, &chat.Algorithm, &chat.User1ID, &chat.User2ID, &chat.CreatedAt, &chat.User2Name); err != nil {
			http.Error(w, "Failed to scan chat", http.StatusInternalServerError)
			return
		}
		chats = append(chats, chat)
	}

	// Проверяем ошибки после итерации
	if err := rows.Err(); err != nil {
		http.Error(w, "Failed to read chats", http.StatusInternalServerError)
		return
	}

	// Возвращаем результат в формате JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(chats); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
