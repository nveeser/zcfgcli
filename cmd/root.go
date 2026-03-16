package cmd

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"time"
	"zcfgcli/sync"
)

type config struct {
	outputDir string
	user      string
	password  string
	broker    string
}

func NewRoot() *cobra.Command {
	var args config

	var rootCmd = &cobra.Command{Use: "zcfgcli"}
	rootCmd.PersistentFlags().StringVar(&args.outputDir, "config", "../devcfg", "Root path to store configuration")
	rootCmd.PersistentFlags().StringVar(&args.user, "user", "", "User name for MQTT server")
	rootCmd.PersistentFlags().StringVar(&args.password, "password", "", "password for MQTT server")
	rootCmd.PersistentFlags().StringVar(&args.broker, "broker", "tcp://10.80.0.100:1883", "MQTT broker address")

	rootCmd.AddCommand(syncCmd{&args}.command())
	rootCmd.AddCommand(applyCmd{&args}.command())
	rootCmd.AddCommand(diffCmd{&args}.command())
	return rootCmd
}

func getSyncer(args *config) *sync.Broker {
	opts := mqtt.NewClientOptions().AddBroker(args.broker)
	opts.Username = args.user
	opts.Password = args.password
	return &sync.Broker{
		Client:  mqtt.NewClient(opts),
		Timeout: 5 * time.Second,
	}
}
