package sync

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"iter"
	"os"
	"zcfgcli/meta"
)

type Device struct {
	descriptor meta.Device
	configText []byte
	config     map[string]any
}

func (d *Device) Name() string            { return d.descriptor.FriendlyName }
func (d *Device) Descriptor() meta.Device { return d.descriptor }
func (d *Device) HasConfig() bool         { return d != nil && len(d.config) > 0 }
func (d *Device) Config() map[string]any  { return d.config }

func FromFile(desc meta.Device, filename string) (*Device, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return &Device{
			descriptor: desc,
			configText: nil,
			config:     make(map[string]any),
		}, nil
	}
	yamlText, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf(" reading %s: %v", filename, err)
	}
	var config map[string]any
	if err := yaml.Unmarshal(yamlText, &config); err != nil {
		return nil, fmt.Errorf("error unmarshaling %s: %v", filename, err)
	}
	return &Device{
		descriptor: desc,
		configText: yamlText,
		config:     config,
	}, nil
}

var cmt = &commentBuilder{60}

func (d *Device) ToNode() (*yaml.Node, error) {
	var configNode yaml.Node
	if err := yaml.Unmarshal(d.configText, &configNode); err != nil {
		return nil, fmt.Errorf("parsing text[%s]: %w", d.Name(), err)
	}
	mapNode := configNode.Content[0]
	mapNode.Style = yaml.LiteralStyle
	var filteredContent []*yaml.Node
	for keyNode, valueNode := range iterMapping(mapNode) {
		entity, ok := d.descriptor.FindEntity(keyNode.Value)
		if !ok || entity.Category != "config" {
			continue
		}
		comment, err := cmt.WriteString(entity)
		if err != nil {
			return nil, err
		}
		keyNode.HeadComment = comment
		filteredContent = append(filteredContent, keyNode, valueNode)
	}
	mapNode.Content = filteredContent
	return &configNode, nil
}

func (d *Device) WriteDescriptor(filename string) error {
	return writeYAML(filename, d.Descriptor())
}

func (d *Device) WriteConfig(filename string) error {
	configNode, err := d.ToNode()
	if err != nil {
		return err
	}
	return writeYAML(filename, configNode)
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
