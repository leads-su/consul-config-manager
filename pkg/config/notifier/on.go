package notifier

// On describes structure of Notifier "on" types
type On struct {
	Error   bool `mapstructure:"error"`
	Success bool `mapstructure:"success"`
}
