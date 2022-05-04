package agent

import "os"

// Network describes structure of agent network configuration
type Network struct {
	Interface string `mapstructure:"interface"`
	Address   string `mapstructure:"address"`
	Port      uint   `mapstructure:"port"`
}

// Hostname returns machine hostname
func (network *Network) Hostname() string {
	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	}
	if len(network.Address) != 0 {
		return network.Address
	}

	if len(network.Interface) != 0 {
		return network.Interface
	}

	return "unknown"
}
