package sync

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v3"
	"iter"
	"log"
	"maps"
	"time"
	"zcfgcli/meta"
)

type Syncer struct {
	mqtt.Client
	deviceMap map[string]*Device
	Timeout   time.Duration
}

type Device struct {
	descriptor meta.Device
	config     map[string]any
	configNode yaml.Node
}

func (c *Device) Descriptor() meta.Device { return c.descriptor }
func (c *Device) ConfigNode() *yaml.Node  { return c.configNode.Content[0] }
func (c *Device) Name() string            { return c.descriptor.FriendlyName }
func (c *Device) HasConfig() bool {
	return !c.configNode.IsZero() &&
		c.configNode.Kind == yaml.DocumentNode &&
		c.configNode.Content[0].Kind == yaml.MappingNode &&
		len(c.configNode.Content[0].Content) > 0
}

func (c *Syncer) Devices() iter.Seq[*Device] { return maps.Values(c.deviceMap) }

func (c *Syncer) Start() error {
	if tk := c.Connect(); tk.Wait() && tk.Error() != nil {
		return tk.Error()
	}
	var devices []meta.Device
	if err := c.pullTopic("zigbee2mqtt/bridge/devices", nil, &devices); err != nil {
		return fmt.Errorf("json.Unmarshal(): %v", err)
	}
	log.Printf("Found %d devices", len(devices))
	c.deviceMap = make(map[string]*Device)
	for _, d := range devices {
		ds := &Device{
			descriptor: d,
		}
		c.deviceMap[d.FriendlyName] = ds
	}
	return nil
}

func (c *Syncer) LoadConfig(d *Device) error {
	name := d.descriptor.FriendlyName
	baseTopic := fmt.Sprintf("zigbee2mqtt/%s", name)
	property, ok := findProperty(d.descriptor)
	if !ok {
		return nil
	}
	whileSub := func() error {
		return c.pushTopic(baseTopic+"/get", map[string]any{property: ""})
	}
	payload, err := c.pullPayload(baseTopic, whileSub)
	if err != nil {
		return fmt.Errorf("loading Device config[%s]: %w", name, err)
	}
	if err := yaml.Unmarshal(payload, &d.configNode); err != nil {
		return fmt.Errorf("loading Device config[%s]: %w", name, err)
	}
	if err := filterAndCommentConfigFields(d); err != nil {
		return fmt.Errorf("adding comments[%s]: %w", name, err)
	}
	return nil
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

//func (c *Syncer) Write(dir string) error {
//	for name, device := range c.deviceMap {
//		basename := friendlyToFile(name)
//		err := writeYAML(path.Join(dir, basename+".meta.yaml"), device.descriptor)
//		if err != nil {
//			return fmt.Errorf("error: %w", err)
//		}
//		err = writeYAML(path.Join(dir, basename+".cfg.yaml"), device.configNode.Content[0])
//		if err != nil {
//			return fmt.Errorf("error: %w", err)
//		}
//	}
//	return nil
//}

func (c *Syncer) pushDevice(friendlyName string, v any) error {
	baseTopic := fmt.Sprintf("zigbee2mqtt/%s", friendlyName)
	return c.pushTopic(baseTopic+"/set", v)
}

func (c *Syncer) pushTopic(topic string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	if tk := c.Publish(topic, 0, true, payload); tk.Wait() && tk.Error() != nil {
		return fmt.Errorf("publishing error: %v", tk.Error())
	}
	return nil
}

func (c *Syncer) pullDevice(friendlyName string, v any) error {
	baseTopic := fmt.Sprintf("zigbee2mqtt/%s", friendlyName)
	return c.pullTopic(baseTopic, nil, v)
}

func (c *Syncer) pullTopic(topic string, f func() error, v any) error {
	got, err := c.pullPayload(topic, f)
	if err != nil {

	}
	if err := json.Unmarshal(got, v); err != nil {
		return fmt.Errorf("json.Unmarshal: into %T: %w", v, err)
	}
	return nil
}

func (c *Syncer) pullPayload(topic string, whileSub func() error) ([]byte, error) {
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

func filterAndCommentConfigFields(device *Device) error {
	cmt := &commentBuilder{60}

	mapNode := device.configNode.Content[0]
	mapNode.Style = yaml.LiteralStyle
	var filteredContent []*yaml.Node
	for keyNode, valueNode := range iterMapping(mapNode) {
		entity, ok := device.descriptor.FindEntity(keyNode.Value)
		if !ok || entity.Category != "config" {
			continue
		}
		comment, err := cmt.WriteString(entity)
		if err != nil {
			return err
		}
		keyNode.HeadComment = comment
		filteredContent = append(filteredContent, keyNode, valueNode)
	}
	mapNode.Content = filteredContent
	return nil
}

func iterMapping(mapping *yaml.Node) iter.Seq2[*yaml.Node, *yaml.Node] {
	if mapping.Kind != yaml.MappingNode {
		panic("oops")
	}
	return func(yield func(*yaml.Node, *yaml.Node) bool) {
		for i := 0; i < len(mapping.Content); i += 2 {
			if !yield(mapping.Content[i], mapping.Content[i+1]) {
				return
			}
		}
	}
}
