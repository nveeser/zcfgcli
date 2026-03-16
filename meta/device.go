package meta

import "strings"

type Device struct {
	FriendlyName       string     `json:"friendly_name"`
	IeeeAddress        string     `json:"ieee_address"`
	Type               string     `json:"type"`
	ModelID            string     `json:"model_id"`
	Manufacturer       string     `json:"manufacturer"`
	DateCode           string     `json:"date_code"`
	Disabled           bool       `json:"disabled"`
	InterviewCompleted bool       `json:"interview_completed"`
	InterviewState     string     `json:"interview_state"`
	Interviewing       bool       `json:"interviewing"`
	NetworkAddress     int        `json:"network_address"`
	PowerSource        string     `json:"power_source"`
	SoftwareBuildID    string     `json:"software_build_id"`
	Supported          bool       `json:"supported"`
	Definition         Definition `json:"definition"`
	Endpoints          map[string]Endpoint
}

func (d Device) Name() string { return d.FriendlyName }

func (d Device) Filename() string {
	return strings.ToLower(strings.Join(strings.Fields(d.FriendlyName), "_"))
}

func (d Device) FindEntity(name string) (Entity, bool) {
	for _, entity := range d.Definition.Exposes {
		if entity.Name == name {
			return entity, true
		}
	}
	return Entity{}, false
}

type Binding struct {
	Cluster string        `json:"cluster"`
	Target  BindingTarget `json:"target"`
}

type BindingTarget struct {
	Endpoint    int    `json:"endpoint"`
	IeeeAddress string `json:"ieee_address"`
	Type        string `json:"type"`
}

type Endpoint struct {
	Bindings []Binding `json:"bindings"`
	Clusters struct {
		Input  []string `json:"input"`
		Output []string `json:"output"`
	} `json:"clusters"`
	ConfiguredReportings []struct {
		Attribute             string  `json:"attribute"`
		Cluster               string  `json:"cluster"`
		MaximumReportInterval int     `json:"maximum_report_interval"`
		MinimumReportInterval int     `json:"minimum_report_interval"`
		ReportableChange      float32 `json:"reportable_change"`
	} `json:"configured_reportings"`
	Name   string `json:"name"`
	Scenes []any  `json:"scenes"`
}
