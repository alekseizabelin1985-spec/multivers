package gameservice

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	server *http.Server
	router *mux.Router
}

func NewHTTPServer(addr string) *HTTPServer {
	router := mux.NewRouter()
	
	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}
	
	return &HTTPServer{
		server: srv,
		router: router,
	}
}

func (hs *HTTPServer) Start() {
	// Запуск HTTP сервера в отдельной горутине
	go func() {
		log.Printf("HTTP server starting on %s", hs.server.Addr)
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
}

func (hs *HTTPServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := hs.server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("HTTP server stopped")
}

func (hs *HTTPServer) RegisterRoutes(service *Service, wsServer *WebSocketServer) {
	// WebSocket endpoints
	hs.router.HandleFunc("/ws/entities", wsServer.HandleWebSocket)
	hs.router.HandleFunc("/ws/events", wsServer.HandleWebSocket)
	hs.router.HandleFunc("/ws/actions", wsServer.HandleWebSocket)
	
	// REST API endpoints
	hs.router.HandleFunc("/entities/{entity_id}", service.GetEntityHandler).Methods("GET")
	hs.router.HandleFunc("/players/register", service.RegisterPlayerHandler).Methods("POST")
	hs.router.HandleFunc("/players/login", service.LoginPlayerHandler).Methods("POST")
	hs.router.HandleFunc("/entities/{entity_id}/history", service.GetEntityHistoryHandler).Methods("GET")
	hs.router.HandleFunc("/events/recent", service.GetRecentEventsHandler).Methods("GET")
}