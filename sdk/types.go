package sdk

import "time"

type SDKConfig struct {
	ProjectID   string       `json:"projectId"`
	Flags       []Flag       `json:"flags"`
	Experiments []Experiment `json:"experiments"`
}

type Flag struct {
	ID      string     `json:"id"`
	Key     string     `json:"key"`
	Enabled bool       `json:"enabled"`
	Rules   []FlagRule `json:"rules"`
}

// FlagRule is a single targeting rule. Value is 0–1 (e.g. 0.3 = 30 %).
type FlagRule struct {
	Type  string  `json:"type"`  // always "percentage" for now
	Value float64 `json:"value"` // 0.0–1.0
}

// The backend only includes running experiments in the config snapshot.
type Experiment struct {
	ID             string    `json:"id"`
	Key            string    `json:"key"`
	TrafficPercent float64   `json:"trafficPercent"` // 0–100
	Variants       []Variant `json:"variants"`
}

type Variant struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Weight int    `json:"weight"`
}

type event struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	ExperimentID string    `json:"experimentId"`
	VariantID    string    `json:"variantId"`
	Type         string    `json:"type"`      // "exposure" | "conversion"
	Name         string    `json:"name"`      // event name, set for conversions
	Value        float64   `json:"value"`
	Timestamp    time.Time `json:"timestamp"`
}

// assignment records a user's variant assignment so Track() can attribute conversions.
type assignment struct {
	ExperimentID string
	VariantID    string
}
