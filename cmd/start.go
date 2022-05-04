package cmd

import (
	"fmt"
	"time"

	"github.com/leads-su/broker"
	"github.com/leads-su/consul-config-manager/pkg/config"
	"github.com/leads-su/consul-config-manager/pkg/http"
	"github.com/leads-su/consul-config-manager/pkg/providers/consul"
	"github.com/leads-su/consul-config-manager/pkg/providers/vault"
	"github.com/leads-su/consul-config-manager/pkg/state"
	"github.com/leads-su/logger"
	"github.com/leads-su/updater"
	"github.com/spf13/cobra"
)

var StartCommand = &cobra.Command{
	Use:   "start",
	Short: "Start CCM",
	Long:  "Start Consul Config Manager",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("cmd:start", "starting application")
		brokerInstance, channel := initializeBroker()
		applicationConfiguration := initializeApplicationConfiguration(brokerInstance)

		logServer := http.NewLogServer(applicationConfiguration)
		logServer.RegisterRoutes()

		eventsServer := http.NewEventServer(applicationConfiguration)
		eventsServer.RegisterRoutes()

		if applicationConfiguration.Updater.Enabled {
			updaterTicker, err := registerUpdateTicker(applicationConfiguration)
			if err != nil {
				logger.Warnf("cmd:start", "failed to register update checker - %s", err.Error())
			}
			defer updaterTicker.Stop()
		}

		if applicationConfiguration.Consul.Enabled {
			go consul.NewConsul(applicationConfiguration)
		}

		if applicationConfiguration.Vault.Enabled {
			go vault.NewVault(applicationConfiguration)
		}

		for {
			switch <-channel {
			case state.ApplicationShutdownRequested:
				fmt.Println("Application shutdown requested")
			case state.ApplicationRestartRequested:
				fmt.Println("Application restart requested")
			case state.ApplicationUpdateRequested:
				fmt.Println("Application update requested")
			case state.ApplicationConfigurationChanged:
				fmt.Println("Application configuration has changed")
			}
		}
	},
}

// initializeBroker initialize broker and return channel
func initializeBroker() (*broker.Broker, chan interface{}) {
	brokerInstance := broker.NewBroker()
	go brokerInstance.Start()
	channel := brokerInstance.Subscribe()
	return brokerInstance, channel
}

// initializeApplicationConfiguration initializes application configuration (default or from file)
func initializeApplicationConfiguration(brokerInstance *broker.Broker) *config.Config {
	appConfig, err := config.Initialize(brokerInstance)
	if err != nil {
		logger.Fatal("cmd:start", "failed to initialize application configuration")
	}
	return appConfig
}

// registerUpdateTicker registers update ticker, so we can now
// check if there is a new version while application is running
func registerUpdateTicker(cfg *config.Config) (*time.Ticker, error) {
	var service updater.UpdaterInterface
	var err error
	updaterTicker := time.NewTicker(60 * time.Minute)

	switch cfg.Updater.Type {
	case "gitlab":
		service, err = updater.InitializeGitlab(updater.GitlabOptions{
			Scheme:      cfg.Updater.Scheme,
			Host:        cfg.Updater.Host,
			Port:        cfg.Updater.Port,
			ApiVersion:  4,
			ProjectID:   cfg.Updater.ProjectID,
			AccessToken: cfg.Updater.AccessToken,
		})
	case "gitea":
		service, err = updater.InitializeGitea(updater.GiteaOptions{
			Scheme:      cfg.Updater.Scheme,
			Host:        cfg.Updater.Host,
			Port:        cfg.Updater.Port,
			Owner:       cfg.Updater.Owner,
			Repository:  cfg.Updater.Repository,
			AccessToken: cfg.Updater.AccessToken,
		})
	default:
		return nil, fmt.Errorf("invalid updater specified - `%s`, only `gitlab` and `gitea` are supported", cfg.Updater.Type)
	}

	if err != nil {
		return nil, err
	}
	service.CheckLatest()
	go func() {
		for {
			select {
			case <-updaterTicker.C:
				service.CheckLatest()
			}
		}
	}()
	return updaterTicker, nil
}
