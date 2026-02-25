package handlers

import (
	"context"
	"io/fs"
	"log"
	"net/http"

	"test-nats/hub"

	"github.com/coder/websocket"
)

func NewMux(uiFS fs.FS, h *hub.Hub) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(uiFS)))
	mux.HandleFunc("/ws", wsHandler(h))
	return mux
}

func wsHandler(h *hub.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Printf("ws accept: %v", err)
			return
		}
		defer conn.CloseNow()

		ctx := r.Context()
		h.Register(ctx, conn)
		defer h.Unregister(conn)

		// Read loop to keep connection alive and detect disconnects
		for {
			_, _, err := conn.Read(ctx)
			if err != nil {
				if ctx.Err() == nil && websocket.CloseStatus(err) != websocket.StatusNormalClosure {
					log.Printf("ws read: %v", err)
				}
				return
			}
		}
	}
}

func GracefulShutdown(ctx context.Context, srv *http.Server) {
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5_000_000_000) // 5s
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown: %v", err)
	}
}
