package config

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Timeout        int            `mapstructure:"timeout"`
	Output         string         `mapstructure:"output"`
	Cache          CacheConfig    `mapstructure:"cache"`
	LookingGlasses []LookingGlass `mapstructure:"looking_glasses"`
}

type CacheConfig struct {
	Enabled bool `mapstructure:"enabled"`
	TTL     int  `mapstructure:"ttl"`
}

type LookingGlass struct {
	Name     string   `mapstructure:"name"`
	URL      string   `mapstructure:"url"`
	Type     string   `mapstructure:"type"`
	Protocol string   `mapstructure:"protocol"`
	Vendor   string   `mapstructure:"vendor"`
	Country  string   `mapstructure:"country"`
	Supports []string `mapstructure:"supports"`
}

func GetDefaultLGs() []LookingGlass {
	return []LookingGlass{
		{
			Name:     "Hurricane Electric",
			URL:      "lg.he.net",
			Type:     "public",
			Protocol: "http",
			Vendor:   "cisco",
			Country:  "US",
			Supports: []string{"bgp", "lookup"},
		},
		{
			Name:     "NTT America",
			URL:      "lg.ntt.net",
			Type:     "public",
			Protocol: "http",
			Vendor:   "cisco",
			Country:  "US",
			Supports: []string{"bgp", "lookup"},
		},
		{
			Name:     "Telia Carrier",
			URL:      "lg.telia.net",
			Type:     "public",
			Protocol: "http",
			Vendor:   "juniper",
			Country:  "SE",
			Supports: []string{"bgp", "lookup"},
		},
	}
}

func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// MarshalJSON marshals data to JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalYAML marshals data to YAML
func MarshalYAML(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}
