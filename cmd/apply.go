package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"path"
	"zcfgcli/sync"
)

type applyCmd struct {
	*config
}

func (c applyCmd) command() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Reads the config files and applies them to the server",
		RunE:  c.Run,
	}
}

func (c *applyCmd) Run(cmd *cobra.Command, args []string) error {
	s := getSyncer(c.config)
	if err := s.Start(); err != nil {
		log.Fatalf("Error starting MQTT syncer: %v", err)
	}

	for _, descriptor := range s.Devices() {
		basename := descriptor.Filename()
		cfgPath := path.Join(c.outputDir, basename+".cfg.yaml")
		localDevice, err := sync.FromFile(descriptor, cfgPath)
		if err != nil {
			return fmt.Errorf("error reading file[%s]: %w", cfgPath, err)
		}
		serverDevice, err := s.PullDevice(descriptor)
		if err != nil {
			return fmt.Errorf("error pulling desc[%s]: %w", cfgPath, err)
		}
		if serverDevice == nil {
			continue
		}
		var updates map[string]any
		log.Printf("%s - Update", descriptor.Name())
		sync.Compare(serverDevice, localDevice, func(key string, before, after any) {
			if updates == nil {
				log.Printf(descriptor.Name())
				updates = make(map[string]any)
			}
			updates[key] = after
		})
		if updates == nil {
			log.Printf("NoOp %s", descriptor.Name())
			continue
		}
		for k, v := range updates {
			config := map[string]any{k: v}
			log.Printf("\t%s %v => %v", k, serverDevice.Config()[k], v)
			if err := s.PushDevice(localDevice.Descriptor(), config); err != nil {
				log.Printf("Error applying config for %s: %v", descriptor.Name(), err)
			}
		}
	}
	return nil
}
