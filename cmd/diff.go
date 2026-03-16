package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"path"
	"zcfgcli/sync"
)

type diffCmd struct {
	*config
}

func (c diffCmd) command() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Compares local config files with the server state",
		RunE:  c.Run,
	}
}

func (c *diffCmd) Run(cmd *cobra.Command, args []string) error {
	s := getSyncer(c.config)
	if err := s.Start(); err != nil {
		return fmt.Errorf("Error starting MQTT syncer: %v", err)
	}

	for _, descriptor := range s.Devices() {
		serverDevice, err := s.PullDevice(descriptor)
		if err != nil {
			log.Printf("Error loading server config for %s: %v", descriptor.Name(), err)
			continue
		}
		if serverDevice == nil {
			continue
		}

		cfgPath := path.Join(c.outputDir, descriptor.Filename()+".cfg.yaml")

		localDevice, err := sync.FromFile(descriptor, cfgPath)
		if err != nil {
			return fmt.Errorf("error loading local config for %w: %v", descriptor.Name(), err)
		}
		log.Printf(descriptor.Name())
		first := true
		sync.Compare(serverDevice, localDevice, func(key string, before, after any) {
			if first {
				log.Printf(descriptor.Name())
				first = false
			}
			log.Printf("\t%s %v => %v", key, before, after)
		})
	}
	return nil
}
