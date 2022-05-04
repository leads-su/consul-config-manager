package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/leads-su/broker"
	"github.com/leads-su/consul-config-manager/pkg/config/agent"
	"github.com/leads-su/consul-config-manager/pkg/config/application"
	"github.com/leads-su/consul-config-manager/pkg/config/consul"
	"github.com/leads-su/consul-config-manager/pkg/config/log"
	"github.com/leads-su/consul-config-manager/pkg/config/notifier"
	"github.com/leads-su/consul-config-manager/pkg/config/sse"
	"github.com/leads-su/consul-config-manager/pkg/config/updater"
	"github.com/leads-su/consul-config-manager/pkg/config/vault"
	"github.com/leads-su/consul-config-manager/pkg/state"
	"github.com/leads-su/logger"
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	Agent       *agent.Agent `mapstructure:"agent"`
	Application *application.Application
	Consul      *consul.Consul     `mapstructure:"consul"`
	Environment string             `mapstructure:"environment"`
	Log         *log.Log           `mapstructure:"log"`
	Sse         *sse.SSE           `mapstructure:"sse"`
	Vault       *vault.Vault       `mapstructure:"vault"`
	Notifier    *notifier.Notifier `mapstructure:"notifier"`
	Updater     *updater.Updater   `mapstructure:"updater"`
}

// Initialize will initialize new instance of config
func Initialize(b *broker.Broker) (*Config, error) {
	config := &Config{
		Agent:       agent.InitializeDefaults(),
		Application: application.InitializeDefaults(),
		Consul:      consul.InitializeDefaults(),
		Environment: "development",
		Log:         log.InitializeDefaults(),
		Sse:         sse.InitializeDefaults(),
		Vault:       vault.InitializeDefaults(),
		Notifier:    notifier.InitializeDefaults(),
		Updater:     updater.InitializeDefaults(),
	}

	err := viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(config.Agent.Network.Address) == "" && strings.TrimSpace(config.Agent.Network.Interface) == "" {
		config.Agent.HealthChecks.HTTP = false
	}

	config.setDefaults()
	config.Log.SetLogLevel()
	config.Log.CollectLogsLocation()
	handleConfigurationUpdate(b)
	return config, nil
}

// setDefaults sets default values which can be used in different parts of application
func (config *Config) setDefaults() {
	//if config.Notifier.Telegram.Token != "" {
	//	viper.SetDefault("notifier.telegram.enabled", true)
	//	viper.SetDefault("notifier.telegram.token", config.Notifier.Telegram.Token)
	//	viper.SetDefault("notifier.telegram.recipients", config.Notifier.Telegram.Recipients)
	//}
}

// handleConfigurationUpdate handle application configuration file update
func handleConfigurationUpdate(b *broker.Broker) {
	viper.WatchConfig()
	viper.OnConfigChange(func(event fsnotify.Event) {
		err := viper.ReadInConfig()
		if err != nil {
			logger.Errorf("config:updater", "failed to re-read configuration file, using previous version - %s", err.Error())
		} else {
			b.Publish(state.ApplicationConfigurationChanged)
		}
	})
}
