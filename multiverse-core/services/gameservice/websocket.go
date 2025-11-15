package gameservice

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Разрешаем подключения с любого источника (в production следует ограничить)
		return true
	},
}

type WebSocketServer struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mutex     sync.Mutex // Мьютекс для синхронизации доступа к clients
}

func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

func (w *WebSocketServer) HandleWebSocket(wr http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(wr, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Регистрируем нового клиента с блокировкой
	w.mutex.Lock()
	w.clients[conn] = true
	w.mutex.Unlock()

	// Обрабатываем входящие сообщения от клиента
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			// Удаляем клиента с блокировкой
			w.mutex.Lock()
			delete(w.clients, conn)
			w.mutex.Unlock()
			break
		}

		// Обрабатываем сообщение от клиента
		// Пытаемся разобрать сообщение как событие
		var event map[string]interface{}
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("Failed to parse client message: %v", err)
			continue
		}

		// TODO: Реализовать обработку действий от клиента
		log.Printf("Received message from client: %v", event)
	}
}

func (w *WebSocketServer) BroadcastMessage(message []byte) {
	// Отправляем сообщение всем подключенным клиентам с блокировкой
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for client := range w.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Failed to send message to client: %v", err)
			client.Close()
			delete(w.clients, client)
		}
	}
}

func (w *WebSocketServer) BroadcastLoop(broadcast <-chan []byte) {
	for message := range broadcast {
		w.BroadcastMessage(message)
	}
}