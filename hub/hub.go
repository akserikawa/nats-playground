package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"test-nats/models"

	"github.com/coder/websocket"
)

const maxRecentResults = 50

type Hub struct {
	mu             sync.RWMutex
	clients        map[*websocket.Conn]struct{}
	recentResults  []models.TaskResult
	consumerStats  map[int]*models.ConsumerStats
	totalPublished int
}

func New() *Hub {
	return &Hub{
		clients:       make(map[*websocket.Conn]struct{}),
		consumerStats: make(map[int]*models.ConsumerStats),
	}
}

func (h *Hub) Register(ctx context.Context, conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}

	state := models.DashboardState{
		Consumers:      h.consumerStats,
		RecentResults:  h.recentResults,
		TotalPublished: h.totalPublished,
	}
	h.mu.Unlock()

	evt := models.WSEvent{
		Type: "initial_state",
		Data: state,
	}
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("hub: marshal initial state: %v", err)
		return
	}
	_ = conn.Write(ctx, websocket.MessageText, data)
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *Hub) Broadcast(evt models.WSEvent) {
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("hub: marshal event: %v", err)
		return
	}

	h.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := c.Write(ctx, websocket.MessageText, data); err != nil {
			cancel()
			h.Unregister(c)
			_ = c.Close(websocket.StatusGoingAway, "write error")
			continue
		}
		cancel()
	}
}

func (h *Hub) IncrementPublished() {
	h.mu.Lock()
	h.totalPublished++
	h.mu.Unlock()
}

func (h *Hub) RecordResult(result models.TaskResult) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.recentResults = append(h.recentResults, result)
	if len(h.recentResults) > maxRecentResults {
		h.recentResults = h.recentResults[len(h.recentResults)-maxRecentResults:]
	}

	stats, ok := h.consumerStats[result.ConsumerID]
	if !ok {
		stats = &models.ConsumerStats{ConsumerID: result.ConsumerID}
		h.consumerStats[result.ConsumerID] = stats
	}
	stats.TotalHandled++
	// Running average
	stats.AvgDurationMs = stats.AvgDurationMs + (float64(result.Duration.Milliseconds())-stats.AvgDurationMs)/float64(stats.TotalHandled)
	stats.LastActive = result.FinishedAt
}
