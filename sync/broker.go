package sync

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v3"
	"log"
	"time"
	"zcfgcli/meta"
)

type Broker struct {
	mqtt.Client
	descriptors []meta.Device
	Timeout     time.Duration
}

func (c *Broker) Devices() []meta.Device { return c.descriptors }

func (c *Broker) Start() error {
	if tk := c.Connect(); tk.Wait() && tk.Error() != nil {
		return tk.Error()
	}
	if err := c.pullTopic("zigbee2mqtt/bridge/devices", nil, &c.descriptors); err != nil {
		return fmt.Errorf("json.Unmarshal(): %v", err)
	}
	log.Printf("Found %d devices", len(c.descriptors))
	return nil
}

func (c *Broker) PullDevice(descriptor meta.Device) (*Device, error) {
	name := descriptor.FriendlyName
	baseTopic := fmt.Sprintf("zigbee2mqtt/%s", name)
	property, ok := findProperty(descriptor)
	if !ok {
		return nil, nil
	}
	whileSub := func() error {
		return c.pushTopic(baseTopic+"/get", map[string]any{property: ""})
	}
	var err error
	configText, err := c.pullPayload(baseTopic, whileSub)
	if err != nil {
		return nil, fmt.Errorf("loading Device config[%s]: %w", name, err)
	}
	d := &Device{
		descriptor: descriptor,
		configText: configText,
	}
	if err := yaml.Unmarshal(configText, &d.config); err != nil {
		return nil, fmt.Errorf("loading Device config[%s]: %w", name, err)
	}
	if err := filterConfigFields(d); err != nil {
		return nil, fmt.Errorf("adding comments[%s]: %w", name, err)
	}
	return d, nil
}

func findProperty(d meta.Device) (string, bool) {
	var props []string
	for _, e := range d.Definition.Exposes {
		if e.Category == "config" {
			props = append(props, e.Property)
		}
		for _, f := range e.Features {
			if f.Property != "" {
				props = append(props, f.Property)
			}
		}
	}
	if len(props) == 0 {
		return "", false
	}
	return props[0], true
}

func (c *Broker) PushDevice(d *Device, config map[string]any) error {
	topic := fmt.Sprintf("zigbee2mqtt/%s/set", d.Name())
	log.Printf("Pushing topic")
	return c.pushTopic(topic, config)
}

func (c *Broker) pushTopic(topic string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	return c.pushPayload(topic, payload)
}

func (c *Broker) pushPayload(topic string, payload []byte) error {
	if tk := c.Publish(topic, 0, true, payload); tk.Wait() && tk.Error() != nil {
		return fmt.Errorf("publishing error: %v", tk.Error())
	}
	return nil
}

func (c *Broker) pullDevice(friendlyName string, v any) error {
	baseTopic := fmt.Sprintf("zigbee2mqtt/%s", friendlyName)
	return c.pullTopic(baseTopic, nil, v)
}

func (c *Broker) pullTopic(topic string, f func() error, v any) error {
	got, err := c.pullPayload(topic, f)
	if err != nil {

	}
	if err := json.Unmarshal(got, v); err != nil {
		return fmt.Errorf("json.Unmarshal: into %T: %w", v, err)
	}
	return nil
}

func (c *Broker) pullPayload(topic string, whileSub func() error) ([]byte, error) {
	if whileSub == nil {
		whileSub = func() error { return nil }
	}
	payloadc := make(chan []byte)
	onMsg := func(c mqtt.Client, m mqtt.Message) {
		payloadc <- m.Payload()
	}
	if st := c.Subscribe(topic, 0, onMsg); st.Wait() && st.Error() != nil {
		return nil, fmt.Errorf("subscribe error: %w", st.Error())
	}
	defer func() {
		if t := c.Unsubscribe(topic); t.Wait() && t.Error() != nil {
			log.Printf("unsubscribe error: %v", t.Error())
		}
		close(payloadc)
	}()

	if err := whileSub(); err != nil {
		return nil, err
	}
	select {
	case got := <-payloadc:
		return got, nil
	case <-time.After(3 * time.Second):
		return nil, fmt.Errorf("Timeout reading payload")
	}
}

func filterConfigFields(device *Device) error {
	for k := range device.config {
		if e, ok := device.descriptor.FindEntity(k); !ok || e.Category != "config" {
			delete(device.config, k)
		}
	}
	return nil
}

//func filterAndCommentConfigFields(device *Device) error {
//	cmt := &commentBuilder{60}
//
//	mapNode := device.configNode.Content[0]
//	mapNode.Style = yaml.LiteralStyle
//	var filteredContent []*yaml.Node
//	for keyNode, valueNode := range iterMapping(mapNode) {
//		entity, ok := device.descriptor.FindEntity(keyNode.Value)
//		if !ok || entity.Category != "config" {
//			continue
//		}
//		comment, err := cmt.WriteString(entity)
//		if err != nil {
//			return err
//		}
//		keyNode.HeadComment = comment
//		filteredContent = append(filteredContent, keyNode, valueNode)
//	}
//	mapNode.Content = filteredContent
//	return nil
//}
