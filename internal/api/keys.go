package api

import (
	"MessengerMPM/internal/db"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func SendPublicKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	var request struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Сохраняем публичный ключ в базе данных
	_, err := db.Pool.Exec(r.Context(), `
        INSERT INTO chat_keys (chat_id, user_id, public_key)
        VALUES ($1, $2, $3)`, chatID, userID, request.PublicKey)
	if err != nil {
		http.Error(w, "Failed to save public key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Public key saved"}`))
}

func GetPublicKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	if chatID == "" {
		http.Error(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int)

	// Получаем ID другого пользователя в чате
	var otherUserID int
	err := db.Pool.QueryRow(r.Context(), `
        SELECT user1_id FROM chats WHERE id = $1 AND user2_id = $2
        UNION
        SELECT user2_id FROM chats WHERE id = $1 AND user1_id = $2`, chatID, userID).Scan(&otherUserID)
	if err != nil {
		http.Error(w, "Failed to find other user", http.StatusInternalServerError)
		return
	}

	// Получаем публичный ключ другого пользователя
	var publicKey string
	err = db.Pool.QueryRow(r.Context(), `
        SELECT public_key FROM chat_keys 
        WHERE chat_id = $1 AND user_id = $2`, chatID, otherUserID).Scan(&publicKey)
	if err != nil {
		http.Error(w, "Failed to get public key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"public_key": "` + publicKey + `"}`))
}
