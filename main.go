package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"test-nats/consumer"
	"test-nats/handlers"
	"test-nats/hub"
	"test-nats/natsserver"
	"test-nats/publisher"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

//go:embed ui
var uiContent embed.FS

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Start embedded NATS server
	ns, clientURL, err := natsserver.Start()
	if err != nil {
		log.Fatalf("nats server: %v", err)
	}
	defer ns.Shutdown()
	log.Printf("NATS server ready at %s", clientURL)

	// 2. Connect client
	nc, err := nats.Connect(clientURL)
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer nc.Close()

	// 3. Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("jetstream: %v", err)
	}

	// 4. Setup stream
	if _, err := publisher.SetupStream(ctx, js); err != nil {
		log.Fatalf("setup stream: %v", err)
	}
	log.Println("JetStream stream TASKS created")

	// 5. Create hub
	h := hub.New()

	// 6. Start HTTP server
	uiFS, err := fs.Sub(uiContent, "ui")
	if err != nil {
		log.Fatalf("ui fs: %v", err)
	}
	mux := handlers.NewMux(uiFS, h)
	srv := &http.Server{Addr: ":8080", Handler: mux}
	go func() {
		log.Println("Dashboard at http://localhost:8080")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()
	go handlers.GracefulShutdown(ctx, srv)

	// 7. Start consumers
	if err := consumer.StartWorkers(ctx, js, 5, h); err != nil {
		log.Fatalf("start workers: %v", err)
	}
	log.Println("5 workers started")

	// 8. Start publisher
	go publisher.Run(ctx, js, h)
	log.Println("Publisher started (800ms interval)")

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down...")
}
