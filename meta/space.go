package meta

import "log"

type Bridge struct {
	Devices []Device
}

func (s *Bridge) Find(name string) (Device, bool) {
	for _, d := range s.Devices {
		if d.FriendlyName == name {
			return d, true
		}
	}
	return Device{}, false
}

func (s *Bridge) showDevices() {
	for _, d := range s.Devices {
		log.Printf("Device: (%s) %s", d.IeeeAddress, d.FriendlyName)
		log.Printf("\t Description: %q", d.Definition.Description)
		for _, e := range d.Definition.Exposes {
			log.Printf("\t Entity(%s): %s(%s) %s [%s]: %s", e.Category, e.Property, e.Type, e.Label, e.Name, e.Description)
			for _, f := range e.Features {
				log.Printf("\t    Feature: %s(%s) [%s]: %s", f.Name, f.Label, f.Type, f.Description)
			}
		}
	}
}
