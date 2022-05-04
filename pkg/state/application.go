package state

const (
	ApplicationConfigurationChanged = iota + applicationIotaValue
	ApplicationUpdateRequested
	ApplicationShutdownRequested
	ApplicationRestartRequested
)
