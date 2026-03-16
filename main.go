package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
	"strings"
	"time"
	"zcfgcli/sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	outputDir = flag.String("config", "../devcfg", "Root path to store configuration")
	host      = flag.String("host", "", "MQTT server host")
	port      = flag.Int("port", 1883, "MQTT server port")
	user      = flag.String("user", "", "User name for MQTT server")
	password  = flag.String("password", "", "password for MQTT server")
)

func main() {
	flag.Parse()
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", *host, *port))
	opts.Username = *user
	opts.Password = *password
	syncer := &sync.Syncer{
		Client:  mqtt.NewClient(opts),
		Timeout: 5 * time.Second,
	}
	if err := syncer.Start(); err != nil {
		log.Fatalf("Error starting MQTT syncer: %v", err)
	}
	var loaded int
	for device := range syncer.Devices() {
		if err := syncer.LoadConfig(device); err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
		if device.HasConfig() {
			loaded++
		} else {
			log.Printf("Skip: %s", device.Name())
		}
	}
	log.Printf("Loaded %d", loaded)
	for d := range syncer.Devices() {
		if !d.HasConfig() {
			continue
		}
		basename := friendlyToFile(d.Name())
		err := writeYAML(path.Join(*outputDir, basename+".meta.yaml"), d.Descriptor())
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		err = writeYAML(path.Join(*outputDir, basename+".cfg.yaml"), d.ConfigNode())
		if err != nil {
			log.Fatalf("error: %v", err)
		}
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

//
//func printMap(w io.Writer, device *device) {
//	for k, v := range device.config {
//		entity, ok := device.descriptor.FindEntity(k)
//		if !ok {
//			log.Printf("could not find entity: %s\n", k)
//			continue
//		}
//		fmt.Fprintf(w, "# --------------------------\n")
//		fmt.Fprintf(w, "# Name: %s\n", entity.Name)
//		fmt.Fprintf(w, "# --------------------------\n")
//		for _, d := range strings.Split(wordwrap.WrapString(entity.Description, 60), "\n") {
//			fmt.Fprintf(w, "# %s\n", d)
//		}
//		fmt.Fprintf(w, "# Type: %s\n", entity.Type)
//		fmt.Fprintf(w, "# Category: %s\n", entity.Category)
//		if entity.Values != nil {
//			fmt.Fprintf(w, "# Values: %s\n", entity.Values)
//		} else {
//			fmt.Fprintf(w, "# Range: (%d, %d)\n", entity.ValueMin, entity.ValueMax)
//		}
//		for _, preset := range entity.Presets {
//			fmt.Fprintf(w, "# Preset: %s => %v %q\n", preset.Name, preset.Value, preset.Description)
//		}
//		fmt.Fprintf(w, "%s: %+v\n", k, v)
//		fmt.Fprintf(w, "\n")
//	}
//}

func limit(d []byte, n int) string {
	dd := string(d)
	if len(dd) > n {
		dd = dd[:n] + "..."
	}
	return dd
}

//	payload := make(chan []byte)
//	tk := client.Subscribe("zigbee2mqtt/#", 0, func(c mqtt.Client, m mqtt.Message) {
//		log.Printf("Received[%s] %s\n", m.Topic(), limit(m.Payload(), 100))
//		payload <- m.Payload()
//	})
//	if tk.Wait() && tk.Error() != nil {
//		log.Fatalf("Error loading config: %v", tk.Error())
//	}
//

//loop:
//	for {
//		select {
//		case <-payload:
//			log.Printf("got")
//		case <-time.After(10 * time.Second):
//			break loop
//		}
//	}
