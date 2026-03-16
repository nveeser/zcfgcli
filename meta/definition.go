package meta

type Definition struct {
	Description string      `json:"description"`
	Model       string      `json:"model"`
	Vendor      string      `json:"vendor"`
	Source      string      `json:"source"`
	SupportsOta bool        `json:"supports_ota"`
	Options     []DefOption `json:"options"`
	Exposes     []Entity    `json:"exposes"`
}

type DefOption struct {
	Access      int     `json:"access"`
	Description string  `json:"description"`
	Label       string  `json:"label"`
	Name        string  `json:"name"`
	Property    string  `json:"property"`
	Type        string  `json:"type"`
	ValueStep   float64 `json:"value_step,omitempty"`
	ValueMax    int     `json:"value_max,omitempty"`
	ValueMin    int     `json:"value_min,omitempty"`
}

type Entity struct {
	Name        string          `json:"name,omitempty"`
	Property    string          `json:"property,omitempty"`
	Label       string          `json:"label,omitempty"`
	Type        string          `json:"type"`
	Unit        string          `json:"unit,omitempty"`
	Access      int             `json:"access,omitempty"`
	Category    string          `json:"category,omitempty"`
	Description string          `json:"description,omitempty"`
	Values      []string        `json:"values,omitempty"`
	ValueMax    int             `json:"value_max,omitempty"`
	ValueMin    int             `json:"value_min,omitempty"`
	Features    []EntityFeature `json:"features,omitempty"`
	Presets     []EntityPreset  `json:"presets,omitempty"`
}

type EntityPreset struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Value       int    `json:"value"`
}

type EntityFeature struct {
	Access      int    `json:"access"`
	Description string `json:"description"`
	Label       string `json:"label"`
	Name        string `json:"name"`
	Property    string `json:"property"`
	Type        string `json:"type"`
	ValueOff    string `json:"value_off,omitempty"`
	ValueOn     string `json:"value_on,omitempty"`
	ValueToggle string `json:"value_toggle,omitempty"`
	ValueMax    int    `json:"value_max,omitempty"`
	ValueMin    int    `json:"value_min,omitempty"`
}
