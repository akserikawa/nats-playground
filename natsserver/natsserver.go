package natsserver

import (
	"fmt"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
)

func Start() (*natsserver.Server, string, error) {
	opts := &natsserver.Options{
		Port:      -1,
		JetStream: true,
		StoreDir:  "",
		NoSigs:    true,
		NoLog:     true,
	}

	ns, err := natsserver.NewServer(opts)
	if err != nil {
		return nil, "", fmt.Errorf("creating nats server: %w", err)
	}

	ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		return nil, "", fmt.Errorf("nats server not ready after 5s")
	}

	clientURL := ns.ClientURL()
	return ns, clientURL, nil
}
