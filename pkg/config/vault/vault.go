package vault

import "fmt"

type Vault struct {
	Enabled    bool       `mapstructure:"enabled"`
	DataCenter string     `mapstructure:"datacenter"`
	Address    *Address   `mapstructure:"address"`
	Addresses  *Addresses `mapstructure:"addresses"`
	Token      string     `mapstructure:"token"`
}

// InitializeDefaults create new vault config instance with default values
func InitializeDefaults() *Vault {
	return &Vault{
		Enabled:    false,
		DataCenter: "dc0",
		Address: &Address{
			Scheme: "http",
			Host:   "127.0.0.1",
			Port:   8200,
		},
		Token: "",
	}
}

// MainAddress returns address for the Vault server we are connecting to
func (vault *Vault) MainAddress() string {
	return fmt.Sprintf("%s:%d", vault.Address.Host, vault.Address.Port)
}
