package notifications

import (
	"time"

	"github.com/dhontecillas/hfw/pkg/ids"
	"github.com/dhontecillas/hfw/pkg/mailer"
)

const (
	// CarrierEmail is to send notifications through email.
	CarrierEmail = "email"
)

// Notification is the templatte of a message to
// be sent to a user
type Notification struct {
	Name string
}

// Carrier is the interface that can be used to send something
type Carrier interface {
	Send() error
}

// Dispatch is a messages sending attempt through a carrier
type Dispatch struct {
	CarrierName            string
	Enqueued               time.Time
	HasDeliverConfirmation bool
	Delivered              time.Time
	Failed                 time.Time
	FailReason             string
}

// Message information about a notification to be sent to a userID
type Message struct {
	ID               ids.ID
	UserID           ids.ID
	NotificationName string
	Enqueued         time.Time
}

// Notifier interface defines the function to send a notification
type Notifier interface {
	Send(notification string, data map[string]interface{}, carriers []string) ([]Message, []error)
}

// Notifications deals with rendering and sending notifications.
type Notifications struct {
	composer Composer
	Mailer   mailer.Mailer
}

// NewNotifications creates a Notifications object to send notifications.
func NewNotifications(composer Composer, mailSender mailer.Mailer) *Notifications {
	return &Notifications{
		composer: composer,
		Mailer:   mailSender,
	}
}

// Send a notification
func (n *Notifications) Send(notification string, data map[string]interface{},
	carriers []string) ([]Message, []error) {

	mailCarrier := NewEmailCarrier(n.Mailer)

	cs, err := n.composer.Render(notification, data, CarrierEmail)
	if err != nil {
		return nil, []error{err}
	}

	err = mailCarrier.Send(cs, data)
	if err != nil {
		return nil, []error{err}
	}

	return nil, nil
}

// NopNotifier implements the Notifier interface doing nothing
type NopNotifier struct {
}

// NewNopNotifier creates a notifier that does nothing.
func NewNopNotifier() *NopNotifier {
	return &NopNotifier{}
}

// Send does nothing at all.
func (n *NopNotifier) Send(notification string, data map[string]interface{},
	carriers []string) ([]Message, []error) {
	return nil, nil
}
