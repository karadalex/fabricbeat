package beater

import (
	"fmt"
	"time"
	"context"
	"io"
	"os"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/karadalex/fabricbeat/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Fabricbeat configuration.
type Fabricbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of fabricbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Fabricbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts fabricbeat.
func (bt *Fabricbeat) Run(b *beat.Beat) error {
	logp.Info("fabricbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		// Get logs from Fabric Docker container
		ctx := context.Background()
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		cli.NegotiateAPIVersion(ctx)

		// Specify containerID from configuration (TODO)
		options := types.ContainerLogsOptions{ShowStdout: true}
		out, err := cli.ContainerLogs(ctx, "6a7bd48821d3", options)
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, out)

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
				"fabric_log": out,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		counter++
	}
}

// Stop stops fabricbeat.
func (bt *Fabricbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
