package consul

import (
	"fmt"
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/leads-su/broker"
	cfg "github.com/leads-su/consul-config-manager/pkg/config"
	"github.com/leads-su/consul-config-manager/pkg/providers/consul/parser"
	"github.com/leads-su/consul-config-manager/pkg/providers/consul/storage"
	consulClient "github.com/leads-su/consul/client"
	consulService "github.com/leads-su/consul/service"
	"github.com/leads-su/consul/state"
	"github.com/leads-su/consul/watcher"
	"github.com/leads-su/logger"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// NewConsul creates new instance of Consul client
func NewConsul(config *cfg.Config) {
	brokerInstance, messageChannel := initializeBroker()
	stopChannel := make(chan bool, 1)
	go run(brokerInstance, messageChannel, config, stopChannel)

	restartRequested := false
	restartInProgress := false

	for {
		switch <-messageChannel {
		case state.ConsulShuttingDown:
			if !restartRequested && !restartInProgress {
				restartRequested = true
			} else if restartRequested && !restartInProgress {
				restartInProgress = true
				go run(brokerInstance, messageChannel, config, stopChannel)
				restartRequested = false
				restartInProgress = false
			}
		case state.ConsulRestartRequested:
			if !restartRequested {
				restartRequested = true
				stopChannel <- true
			}
		}
	}
}

// run initializes connection to Consul
func run(brokerInstance *broker.Broker, messageChannel chan interface{}, config *cfg.Config, stopChannel chan bool) {
	client := createClientConfiguration(brokerInstance, messageChannel, config)
	client.SelectBestServer().Connect()

	service := registerService(config, client)

	updateChannel := make(chan consulAPI.KVPairs)
	errorChannel := make(chan error)

	consulWatcher := &watcher.Watcher{
		Client:        client.APIClient(),
		Prefix:        "/",
		UpdateChannel: updateChannel,
		ErrorChannel:  errorChannel,
	}
	consulParser := parser.NewParser()
	consulStorage := storage.NewStorage(config, consulParser)

	go consulWatcher.Start()
	defer consulWatcher.Stop()

	for {
		select {
		case values := <-updateChannel:
			consulParser.ProcessReceivedData(values)
			consulStorage.ProcessChanges(consulParser.GenerateConfiguration())
		case err := <-errorChannel:
			fmt.Printf("%s\n", err.Error())
		case <-stopChannel:
			deregisterService(service)
			brokerInstance.Publish(state.ConsulShuttingDown)
			return
		}
	}
}

// createClientConfiguration creates new instance of Consul configuration
func createClientConfiguration(brokerInstance *broker.Broker, messageChannel chan interface{}, config *cfg.Config) *consulClient.Client {
	var connections []*consulClient.ConnectionInformation

	for _, address := range config.Consul.Addresses {
		connection := consulClient.NewConnection(&consulClient.Connection{
			Scheme:      address.Scheme,
			Host:        address.Host,
			Port:        address.Port,
			DataCenter:  config.Consul.DataCenter,
			AccessToken: config.Consul.Token,
		})
		connections = append(connections, connection)
	}

	return consulClient.WithCustomBroker(brokerInstance, messageChannel).MultipleServers(connections)
}

// initializeBroker returns instance of broker and channel
func initializeBroker() (*broker.Broker, chan interface{}) {
	instance := broker.NewBroker()
	go instance.Start()
	channel := instance.Subscribe()
	return instance, channel
}

// registerService register consul service
func registerService(config *cfg.Config, client *consulClient.Client) *consulService.Service {
	service := consulService.NewService(consulService.Options{
		Client:     client,
		Name:       "ccm",
		Scheme:     "http",
		Host:       config.Agent.Address(),
		Port:       config.Agent.Network.Port,
		HttpServer: config.Agent.HealthChecks.HTTP,
		Interval:   time.Second * 10,
		Timeout:    time.Second * 30,
		ExtraMeta: map[string]string{
			"config_name": viper.GetString("application.configuration_file"),
			"config_path": viper.GetString("application.configuration_file_path"),
			"log_path":    viper.GetString("application.log_path"),
			"log_level":   viper.GetString("application.log_level"),
			"environment": config.Environment,
		},
	})
	err := service.Register()
	if err != nil {
		logger.Fatalf("consul:service", "%s", err.Error())
	}
	setupProviderShutdownHandler(service)
	return service
}

// deregisterService deregister consul service
func deregisterService(service *consulService.Service) {
	err := service.Deregister()
	if err != nil {
		logger.Errorf("consul:service", "%s", err.Error())
	}
}

// setupProviderShutdownHandler initialize provider shutdown handler
func setupProviderShutdownHandler(service *consulService.Service) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChannel
		service.Deregister()
		os.Exit(0)
	}()
}
