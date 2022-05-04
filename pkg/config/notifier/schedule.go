package notifier

// Schedule describes structure of Notifier schedule
type Schedule struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
	Token    string `mapstructure:"token"`
}
