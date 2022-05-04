package notifier

import (
	"github.com/leads-su/consul-config-manager/pkg/config/notifier/notifiers"
	"github.com/leads-su/logger"
	notifierPackage "github.com/leads-su/notifier"
)

const (
	TELEGRAM = 0
)

// Notifier describes structure of Notifier
type Notifier struct {
	Enabled   bool                 `mapstructure:"enabled"`
	Schedule  *Schedule            `mapstructure:"schedule"`
	Notifiers *notifiers.Notifiers `mapstructure:"notifiers"`
	NotifyOn  *On                  `mapstructure:"notify_on"`
}

// InitializeDefaults create new notifier config instance with default values
func InitializeDefaults() *Notifier {
	return &Notifier{
		Enabled: false,
		Schedule: &Schedule{
			Enabled:  false,
			Endpoint: "",
			Token:    "",
		},
		Notifiers: nil,
		NotifyOn: &On{
			Error:   true,
			Success: false,
		},
	}
}

// DeliverNotification delivers notification to a given service
func (notifier *Notifier) DeliverNotification(service int, notification *notifierPackage.Notification) {
	if notifier.IsEnabled() {
		switch service {
		case TELEGRAM:
			go func() {
				err := notifier.Notifiers.GetTelegram().DeliverNotification(
					notification,
					notifier.Notifiers.Telegram.GetRecipients(),
				)
				if err != nil {
					logger.Errorf("notifier:telegram", "failed to deliver notification - %s", err.Error())
				}
			}()
		default:
			logger.Errorf("notifier", "unknown service with ID - %d", service)
		}
	} else {
		logger.Warn("notifier", "tried to send notification while notifier service is disabled")
	}
}

// IsEnabled checks if notifier is enabled
func (notifier *Notifier) IsEnabled() bool {
	if !notifier.Enabled {
		return false
	}
	// For now, only Telegram is important.
	// So, if Telegram is disabled, there is no reason to enable notifier.
	return notifier.Notifiers.Telegram.IsEnabled()
}

// HasNotifiers checks if notifier has configured notifiers
func (notifier *Notifier) HasNotifiers() bool {
	return notifier.Notifiers != nil
}

// HasSchedule checks if notifier has schedule enabled
func (notifier *Notifier) HasSchedule() bool {
	return notifier.Schedule.Enabled
}
