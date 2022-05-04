package consul

type Consul struct {
	Enabled    bool   `mapstructure:"enabled"`
	DataCenter string `mapstructure:"datacenter"`
	Address    *Address
	Addresses  Addresses `mapstructure:"addresses"`
	Token      string    `mapstructure:"token"`
	WriteTo    string    `mapstructure:"write_to"`
}

// InitializeDefaults create new consul config instance with default values
func InitializeDefaults() *Consul {
	return &Consul{
		Enabled:    true,
		DataCenter: "dc0",
		Addresses: Addresses{
			&Address{
				Scheme: "http",
				Host:   "127.0.0.1",
				Port:   8500,
			},
		},
		Token:   "",
		WriteTo: "/etc/ccm.d",
	}
}
