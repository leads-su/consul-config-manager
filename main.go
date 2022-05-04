package main

import (
	"github.com/leads-su/application"
	"github.com/leads-su/application/commands"
	"github.com/leads-su/consul-config-manager/cmd"
	"github.com/leads-su/logger"
)

func main() {
	app := application.NewApplication(application.Options{
		ShortDescription:       "Application Configuration Manager with Consul as a backend",
		LongDescription:        "Manage configuration for applications using Consul as a backend for KV storage",
		HasConfiguration:       true,
		ConfigurationPath:      "/etc/ccm.d",
		ConfigurationFile:      "config.yml",
		ConfigurationEnvPrefix: "CCM",
	})
	app.RegisterCommand(cmd.StartCommand)
	app.RegisterCommand(commands.VersionCommand)

	err := app.Start()
	if err != nil {
		logger.Fatalf("main", "failed to initialize application - %s", err.Error())
	}
}
