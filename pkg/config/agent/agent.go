package agent

import (
	"fmt"
	"github.com/leads-su/network"
	"strings"
)

type Agent struct {
	Network      *Network      `mapstructure:"network"`
	HealthChecks *HealthChecks `mapstructure:"health_check"`
}

// InitializeDefaults create new agent config instance with default values
func InitializeDefaults() *Agent {
	return &Agent{
		Network: &Network{
			Interface: "",
			Address:   "",
			Port:      32175,
		},
		HealthChecks: &HealthChecks{
			TTL:  true,
			HTTP: false,
		},
	}
}

// Address returns address string for Local Agent
func (agent *Agent) Address() string {
	definedAddress := strings.TrimSpace(agent.Network.Address)
	if definedAddress != "" {
		return definedAddress
	}

	ipAddress, err := network.GetIPv4ByName(agent.Network.Interface)
	if err != nil {
		agent.HealthChecks.HTTP = false
	}
	return ipAddress
}

// AddressWithPort returns address string with port for Local Agent
func (agent *Agent) AddressWithPort() string {
	ipAddress := agent.Address()
	return fmt.Sprintf("%s:%d", ipAddress, agent.Network.Port)
}
