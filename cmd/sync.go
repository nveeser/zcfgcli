package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"zcfgcli/sync"
)

type syncCmd struct {
	*config
}

func (c syncCmd) command() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Reads the server and updates the yaml files",
		RunE:  c.Run,
	}
}

func (c *syncCmd) Run(cmd *cobra.Command, args []string) error {
	s := getSyncer(c.config)
	if err := s.Start(); err != nil {
		log.Fatalf("Error starting MQTT syncer: %v", err)
	}
	devices := make(map[string]*sync.Device)
	for _, desc := range s.Devices() {
		serverDevice, err := s.PullDevice(desc)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
		if serverDevice != nil {
			devices[serverDevice.Name()] = serverDevice
		}
	}
	log.Printf("Loaded %d", len(devices))
	for _, serverDevice := range devices {
		basename := serverDevice.Descriptor().Filename()
		basedir := path.Dir(path.Join(c.outputDir, basename))
		if err := os.MkdirAll(basedir, 0755); err != nil {
			return fmt.Errorf("error making directories: %w", err)
		}

		err := serverDevice.WriteDescriptor(path.Join(c.outputDir, basename+".meta.yaml"))
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		err = serverDevice.WriteConfig(path.Join(c.outputDir, basename+".cfg.yaml"))
		if err != nil {
			log.Fatalf("error: %v", err)
		}
	}
	return nil
}
