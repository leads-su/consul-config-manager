package agent

type HealthChecks struct {
	TTL  bool `mapstructure:"ttl"`
	HTTP bool `mapstructure:"http"`
}
