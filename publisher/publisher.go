package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"test-nats/hub"
	"test-nats/models"

	"github.com/nats-io/nats.go/jetstream"
)

func SetupStream(ctx context.Context, js jetstream.JetStream) (jetstream.Stream, error) {
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "TASKS",
		Subjects:  []string{"tasks.work"},
		Retention: jetstream.WorkQueuePolicy,
		Storage:   jetstream.MemoryStorage,
	})
	if err != nil {
		return nil, fmt.Errorf("creating stream: %w", err)
	}
	return stream, nil
}

func Run(ctx context.Context, js jetstream.JetStream, h *hub.Hub) {
	ticker := time.NewTicker(800 * time.Millisecond)
	defer ticker.Stop()

	var counter atomic.Int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			id := int(counter.Add(1))
			task := models.Task{
				ID:        id,
				Payload:   fmt.Sprintf("task-payload-%d", id),
				CreatedAt: time.Now(),
			}

			data, err := json.Marshal(task)
			if err != nil {
				log.Printf("marshal task: %v", err)
				continue
			}

			if _, err := js.Publish(ctx, "tasks.work", data); err != nil {
				log.Printf("publish task %d: %v", id, err)
				continue
			}

			h.Broadcast(models.WSEvent{
				Type: "task_published",
				Data: task,
			})
			h.IncrementPublished()

			log.Printf("published task-%d", id)
		}
	}
}
