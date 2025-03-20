package main

import (
	"MessengerMPM/internal/api"
	"MessengerMPM/internal/db"
	"MessengerMPM/internal/rabbitmq"
	"MessengerMPM/internal/websocket"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Инициализация RabbitMQ
	rabbit, err := rabbitmq.New("amqp://user:password@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbit.Close()
	log.Println("Connected to RabbitMQ")

	connString := "postgres://user_ms:password_ms@postgres:5432/ms_db"
	if err := db.InitDB(connString); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to messenger database")

	// Инициализация WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// Инициализация API с зависимостями
	api.InitAPI(hub, rabbit)

	// Создаем новый маршрутизатор
	r := mux.NewRouter()

	// Маршруты для аутентификации
	r.HandleFunc("/auth/register", api.Register).Methods("POST")
	r.HandleFunc("/auth/login", api.Login).Methods("POST")
	r.HandleFunc("/auth/logout", api.Logout).Methods("POST")

	// Защищенные маршруты для чатов
	protected := r.PathPrefix("/chats").Subrouter()
	protected.Use(api.AuthMiddleware)
	protected.HandleFunc("", api.CreateChat).Methods("POST")
	protected.HandleFunc("", api.GetUserChats).Methods("GET")
	protected.HandleFunc("/{chat_id}", api.DeleteChat).Methods("DELETE")
	protected.HandleFunc("/{chat_id}/keys", api.SendPublicKey).Methods("POST")
	protected.HandleFunc("/{chat_id}/keys", api.GetPublicKey).Methods("GET")
	protected.HandleFunc("/{chat_id}/messages", api.SendMessage).Methods("POST")
	protected.HandleFunc("/{chat_id}/messages", api.GetMessages).Methods("GET")

	r.HandleFunc("/ws", hub.HandleWebSocket)

	// Настройка CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Разрешить все origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"}, // Разрешить все заголовки
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
