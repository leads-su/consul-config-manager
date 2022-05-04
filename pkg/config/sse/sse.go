package sse

// SSE describes structure for `sse` configuration section
type SSE struct {
	WriteTo string `mapstructure:"write_to"`
}

// InitializeDefaults create new log config instance with default values
func InitializeDefaults() *SSE {
	return &SSE{
		WriteTo: "/var/log/ccm/events",
	}
}
