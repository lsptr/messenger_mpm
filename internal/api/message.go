package api

import (
	"MessengerMPM/internal/db"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
)

func SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		// EncryptedMessage string `json:"encrypted_message"`
		Content   string `json:"content"`
		FileData  string `json:"file"`
		FileName  string `json:"file_name"`
		Algorithm string `json:"algorithm"`
		Key       string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Получаем имя отправителя
	var senderName string
	err := db.Pool.QueryRow(r.Context(), `
        SELECT username FROM users WHERE id = $1`, userID).Scan(&senderName)
	if err != nil {
		http.Error(w, "Failed to get sender name", http.StatusInternalServerError)
		return
	}

	// Проверяем, что chat_id и user_id существуют
	var chatExists, userExists bool
	err = db.Pool.QueryRow(r.Context(), `
        SELECT EXISTS(SELECT 1 FROM chats WHERE id = $1), 
               EXISTS(SELECT 1 FROM users WHERE id = $2)`, chatID, userID).Scan(&chatExists, &userExists)
	if err != nil {
		http.Error(w, "Failed to validate chat or user", http.StatusInternalServerError)
		return
	}
	if !chatExists || !userExists {
		http.Error(w, "Chat or user does not exist", http.StatusBadRequest)
		return
	}

	// Получаем информацию о чате
	var user1ID, user2ID int
	err = db.Pool.QueryRow(r.Context(), `
        SELECT user1_id, user2_id FROM chats WHERE id = $1`, chatID).Scan(&user1ID, &user2ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Chat not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get chat details", http.StatusInternalServerError)
		}
		return
	}

	// Определяем, кому адресовано сообщение
	var toUserID int
	if userID == user1ID {
		toUserID = user2ID
	} else if userID == user2ID {
		toUserID = user1ID
	} else {
		http.Error(w, "You are not a member of this chat", http.StatusForbidden)
		return
	}

	// Шифруем сообщение
	var encryptedMessage []byte
	switch request.Algorithm {
	case "RC5":
		encryptedMessage, err = EncryptRC5([]byte(request.Content), []byte(request.Key))
		log.Printf("encryptedMessage: %x", encryptedMessage)
	case "Serpent":
		encryptedMessage, err = EncryptSerpent([]byte(request.Content), []byte(request.Key))
		log.Printf("encryptedMessage: %x", encryptedMessage)
	default:
		http.Error(w, "Unsupported algorithm", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "Failed to encrypt message", http.StatusInternalServerError)
		return
	}

	// Сохраняем метаданные сообщения
	var messageID int
	var createdAt time.Time
	err = db.Pool.QueryRow(r.Context(), `
        INSERT INTO messages (chat_id, user_id)
        VALUES ($1, $2)
        RETURNING id, created_at`, chatID, userID).Scan(&messageID, &createdAt)
	if err != nil {
		log.Printf("Ошибка при сохранении метаданных сообщения: %v", err)
		http.Error(w, "Failed to save message metadata", http.StatusInternalServerError)
		return
	}

	log.Printf("Сообщение успешно сохранено: ID=%d, created_at=%v", messageID, createdAt)

	encryptedMessageBase64 := base64.StdEncoding.EncodeToString(encryptedMessage)
	log.Printf("encryptedMessageBase64: %x", encryptedMessageBase64)
	// Формируем данные для RabbitMQ
	messageData := map[string]interface{}{
		"id":         messageID,
		"senderID":   userID,
		"senderName": senderName,
		"receiverID": toUserID,
		"file": map[string]string{
			"data": request.FileData,
			"name": request.FileName,
		},
		"message":   encryptedMessageBase64,
		"createdAt": createdAt.Format(time.RFC3339),
	}
	messageBytes, err := json.Marshal(messageData)
	if err != nil {
		log.Printf("Ошибка при маршалинге сообщения: %v", err)
		http.Error(w, "Failed to marshal message", http.StatusInternalServerError)
		return
	}

	// Отправляем сообщение в RabbitMQ
	queueName := "chat_" + chatID + "_user_" + strconv.Itoa(toUserID)
	err = rabbit.Publish(queueName, messageBytes)
	if err != nil {
		http.Error(w, "Failed to send message via RabbitMQ", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Message sent"}`))
}

//Получение сообщений

func GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Проверяем, что пользователь является участником чата
	var userName string
	err := db.Pool.QueryRow(r.Context(), `
        SELECT username
		FROM users u
		WHERE u.id = $1
        `, userID).Scan(&userName)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, "Failed to check chat membership", http.StatusInternalServerError)
		return
	}

	var exists bool
	err = db.Pool.QueryRow(r.Context(), `
        SELECT EXISTS(
            SELECT 1 FROM chats 
            WHERE id = $1 AND (user1_id = $2 OR user2_id = $2)
        )`, chatID, userID).Scan(&exists)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, "Failed to check chat membership", http.StatusInternalServerError)
		return
	}
	if !exists {
		log.Printf("error: %v", err)
		http.Error(w, "You are not a member of this chat", http.StatusForbidden)
		return
	}

	// Получаем ключ и алгоритм из запроса
	key := r.Header.Get("X-Key")
	algorithm := r.Header.Get("X-Algorithm")
	log.Printf(key)
	log.Printf(algorithm)

	// Получаем сообщения из RabbitMQ
	queueName := "chat_" + chatID + "_user_" + strconv.Itoa(userID)
	messages, err := rabbit.GetMessages(queueName, 100) // Получаем до 100 сообщений
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, "Failed to get messages from RabbitMQ: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if messages == nil {
		messages = []map[string]interface{}{}
	}

	// Дешифруем сообщения
	for i, msg := range messages {
		encryptedMessageBase64 := msg["message"].(string)
		encryptedMessage, err := base64.StdEncoding.DecodeString(encryptedMessageBase64)
		log.Printf("encryptedMessageBase64 from rMQ %x", encryptedMessageBase64)
		log.Printf("encryptedMessage from rMQ %x", encryptedMessage) // Выведет шифртекст в виде HEX
		var decryptedMessage []byte
		switch algorithm {
		case "RC5":
			decryptedMessage, err = DecryptRC5([]byte(encryptedMessage), []byte(key))
			log.Printf("decryptedMessage %x", decryptedMessage)
		case "Serpent":
			decryptedMessage, err = DecryptSerpent([]byte(encryptedMessage), []byte(key))
		default:
			http.Error(w, "Unsupported algorithm", http.StatusBadRequest)
			log.Printf("decryptedMessage %x", decryptedMessage)
			return
		}
		if err != nil {
			http.Error(w, "Failed to decrypt message", http.StatusInternalServerError)
			return
		}
		messages[i]["message"] = string(decryptedMessage)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"userName": userName,
	})
}
