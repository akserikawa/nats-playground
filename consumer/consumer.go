package consumer

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"test-nats/hub"
	"test-nats/models"

	"github.com/nats-io/nats.go/jetstream"
)

func StartWorkers(ctx context.Context, js jetstream.JetStream, n int, h *hub.Hub) error {
	cons, err := js.CreateOrUpdateConsumer(ctx, "TASKS", jetstream.ConsumerConfig{
		Durable:       "workers",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "tasks.work",
	})
	if err != nil {
		return err
	}

	for i := 1; i <= n; i++ {
		go worker(ctx, cons, i, h)
	}

	return nil
}

func worker(ctx context.Context, cons jetstream.Consumer, id int, h *hub.Hub) {
	log.Printf("worker-%d started", id)
	for {
		select {
		case <-ctx.Done():
			log.Printf("worker-%d stopping", id)
			return
		default:
		}

		msgs, err := cons.Fetch(1, jetstream.FetchMaxWait(2*time.Second))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		for msg := range msgs.Messages() {
			var task models.Task
			if err := json.Unmarshal(msg.Data(), &task); err != nil {
				log.Printf("worker-%d: unmarshal: %v", id, err)
				_ = msg.Nak()
				continue
			}

			// Simulate processing: 50-500ms
			processDuration := time.Duration(50+rand.Intn(451)) * time.Millisecond
			time.Sleep(processDuration)

			if err := msg.Ack(); err != nil {
				log.Printf("worker-%d: ack task-%d: %v", id, task.ID, err)
				continue
			}

			result := models.TaskResult{
				TaskID:     task.ID,
				ConsumerID: id,
				Status:     "completed",
				Duration:   processDuration,
				FinishedAt: time.Now(),
			}

			h.RecordResult(result)
			h.Broadcast(models.WSEvent{
				Type: "task_completed",
				Data: result,
			})

			log.Printf("worker-%d processed task-%d (%s)", id, task.ID, processDuration)
		}

		if err := msgs.Error(); err != nil {
			if ctx.Err() != nil {
				return
			}
		}
	}
}
