package cmd

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"os"
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
	rootCmd.PersistentFlags().StringVar(&args.broker, "broker",
		os.Getenv("Z2M_BROKER"), "MQTT broker address")
	rootCmd.PersistentFlags().StringVar(&args.user, "user",
		os.Getenv("Z2M_USER"), "User name for MQTT server")
	rootCmd.PersistentFlags().StringVar(&args.password, "password",
		os.Getenv("Z2M_PASSWORD"), "password for MQTT server")
	rootCmd.PersistentFlags().StringVar(&args.outputDir, "config",
		os.Getenv("Z2M_CONFIG"), "Root path to store configuration")

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
