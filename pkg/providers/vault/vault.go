package vault

import (
	cfg "github.com/leads-su/consul-config-manager/pkg/config"
	"github.com/leads-su/logger"
)

// NewVault creates new instance of Vault client
func NewVault(config *cfg.Config) {
	logger.Warnf("providers:vault", "vault provider is not implemented")
}
