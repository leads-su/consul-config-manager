package notifiers

import (
	"strconv"
)

// TelegramNotifierConfiguration describes structure for Telegram notifier
type TelegramNotifierConfiguration struct {
	Enabled    bool   `mapstructure:"enabled"`
	Token      string `mapstructure:"token"`
	Recipients []int  `mapstructure:"recipients"`
}

// IsEnabled returns Telegram notifier activation status
func (tn *TelegramNotifierConfiguration) IsEnabled() bool {
	if len(tn.Recipients) == 0 {
		return false
	}
	return tn.Enabled
}

// GetToken returns Telegram token used to authenticate bot
func (tn *TelegramNotifierConfiguration) GetToken() string {
	return tn.Token
}

// GetRecipients returns list of recipients for Telegram messages
func (tn *TelegramNotifierConfiguration) GetRecipients() []string {
	var recipients []string

	for _, recipient := range tn.Recipients {
		recipients = append(recipients, strconv.Itoa(recipient))
	}

	return recipients
}
