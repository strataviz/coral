package buildqueue

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

const (
	DefaultNatsMonitorPort = 8222
	DefaultNatsPort        = 4222
	DefaultNatsClusterPort = 6222

	DefaultLameDuckDuration    = "30s"
	DefaultLameDuckGracePeriod = "10s"
	DefaultPidFile             = "/var/run/nats/nats.pid"
)

type NatsClusterConfig struct {
	Name        string   `json:"name"`
	NoAdvertise bool     `json:"no_advertise"`
	Port        int      `json:"port"`
	Routes      []string `json:"routes"`
}

type NatsConfig struct {
	HttpPort            int                `json:"http_port"`
	LameDuckDuration    string             `json:"lame_duck_duration"`
	LameDuckGracePeriod string             `json:"lame_duck_grace_period"`
	PidFile             string             `json:"pid_file"`
	Port                int                `json:"port"`
	Cluster             *NatsClusterConfig `json:"cluster,omitempty"`
}

func NewNatsConfig(name string, replicas int32) *NatsConfig {
	config := &NatsConfig{
		HttpPort:            DefaultNatsMonitorPort,
		LameDuckDuration:    DefaultLameDuckDuration,
		LameDuckGracePeriod: DefaultLameDuckGracePeriod,
		PidFile:             DefaultPidFile,
		Port:                DefaultNatsPort,
	}

	if replicas > 1 {
		config.Cluster = &NatsClusterConfig{
			Name:        name,
			NoAdvertise: false,
			Port:        DefaultNatsClusterPort,
		}

		for i := 0; i < int(replicas); i++ {
			config.Cluster.Routes = append(
				config.Cluster.Routes,
				fmt.Sprintf("nats://%s-nats-%d.%s-nats-headless:%d", name, i, name, DefaultNatsClusterPort),
			)
		}
	}

	return config
}

func (c *NatsConfig) GetConfigMapData() map[string]string {
	data := c.ToBytes()
	return map[string]string{
		"nats.conf": string(data),
	}
}

func (c *NatsConfig) ToBytes() []byte {
	data, err := json.Marshal(c)
	if err != nil {
		// TODO: something else.  we should never actually be in a position
		// where we can't marshal the config since there's no user input, but
		// we should still handle this case.
		return []byte{}
	}

	return data
}

func (c *NatsConfig) Checksum() string {
	data := c.ToBytes()
	sum := sha256.Sum256(data)
	return string(sum[:])
}
