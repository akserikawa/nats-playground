# nats-playground

A single-binary Go playground that demonstrates NATS queue group load distribution with a real-time WebSocket dashboard. Run `go run .` and open the browser — no external dependencies needed (embedded NATS server).

## Architecture

```
publisher (800ms ticker)
    │
    ▼
NATS JetStream Stream "TASKS" (WorkQueue retention, memory storage)
    │
    ├── worker-1 ──┐
    ├── worker-2 ──┤
    ├── worker-3 ──┼──► hub.Broadcast() ──► WebSocket ──► Browser Dashboard
    ├── worker-4 ──┤
    └── worker-5 ──┘
```

- **Publisher** — publishes a task every 800ms to `tasks.work`
- **Consumers** — 5 goroutines sharing one JetStream durable consumer with simulated 50–500ms processing
- **Hub** — fan-out to WebSocket clients, tracks state for late-joining browsers
- **Dashboard** — 4-panel dark-theme UI (live feed, bar chart, stats, timeline)

## Quick Start

```sh
go run .
```

Open [http://localhost:8080](http://localhost:8080). Ctrl+C to stop.

## Dashboard

![dashboard](https://github.com/user-attachments/assets/placeholder)

| Panel | Description |
|-------|-------------|
| Live Feed | Scrolling log of task completions, color-coded by worker |
| Distribution | Horizontal bars showing task count per worker |
| Stats | Total published, completed, pending, and average latency |
| Timeline | Colored dots showing the assignment pattern over time |

## Project Structure

```
├── main.go              # Orchestrator: startup, graceful shutdown
├── models/models.go     # Shared types (Task, TaskResult, WSEvent, etc.)
├── natsserver/           # Embedded NATS server with JetStream
├── publisher/            # Stream setup + ticker-based publishing
├── consumer/             # N worker goroutines with Fetch(1) loop
├── hub/                  # WebSocket fan-out hub with state tracking
├── handlers/             # HTTP mux: embedded UI + /ws endpoint
└── ui/index.html         # Self-contained dashboard (HTML+CSS+JS)
```

## Dependencies

- [nats.go](https://github.com/nats-io/nats.go) — NATS client + JetStream API
- [nats-server](https://github.com/nats-io/nats-server) — embedded server
- [coder/websocket](https://github.com/coder/websocket) — WebSocket server
