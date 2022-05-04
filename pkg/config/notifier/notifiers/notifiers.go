package notifiers

import (
	"github.com/leads-su/notifier"
)

// Notifiers describes notifiers structure
type Notifiers struct {
	Telegram *TelegramNotifierConfiguration `mapstructure:"telegram"`
}

// GetTelegram returns configuration for Telegram notifier
func (notifiers *Notifiers) GetTelegram() *notifier.TelegramNotifier {
	return notifier.NewTelegramNotifier(notifiers.Telegram.Token)
}
