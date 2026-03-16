package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"zcfgcli/cmd"
)

func main() {
	rootCmd := cmd.NewRoot()
	//var args config
	//
	//var rootCmd = &cobra.Command{Use: "zcfgcli"}
	//rootCmd.PersistentFlags().StringVar(&args.outputDir, "config", "../devcfg", "Root path to store configuration")
	//rootCmd.PersistentFlags().StringVar(&args.user, "user", "", "User name for MQTT server")
	//rootCmd.PersistentFlags().StringVar(&args.password, "password", "", "password for MQTT server")
	//rootCmd.PersistentFlags().StringVar(&args.broker, "broker", "tcp://10.80.0.100:1883", "MQTT broker address")
	//
	//rootCmd.AddCommand(syncCmd{&args}.command())
	//rootCmd.AddCommand(applyCmd{&args}.command())
	//rootCmd.AddCommand(diffCmd{&args}.command())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func writeYAML(filename string, v any) error {
	yamlData, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("yaml.Marshal[%s]: %w", filename, err)
	}
	if err := os.WriteFile(filename, yamlData, 0644); err != nil {
		return fmt.Errorf("file error: %w", err)
	}
	return nil
}

func friendlyToFile(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), "_"))
}

func limit(d []byte, n int) string {
	dd := string(d)
	if len(dd) > n {
		dd = dd[:n] + "..."
	}
	return dd
}
