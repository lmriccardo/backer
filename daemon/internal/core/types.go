package core

import (
	"encoding/json"
	"fmt"
)

type JobStatus int

const (
	JobStatusDisabled JobStatus = iota
	JobStatusEnabled
	JobStatusAll
)

type NotificationType string

const (
	EmailNotify   NotificationType = "email"
	WebhookNotify NotificationType = "webhook"
)

type Webhook string

const (
	WebhookDiscord Webhook = "discord"
)

type EventType string

const (
	EventFailure EventType = "failure"
	EventSuccess EventType = "success"
)

type Notification interface {
	Kind() string // Returns the notification system type
}

type NotificationList []Notification

type JobConfig struct {
	Name        string           `json:"name"`         // The name of the job
	Log         string           `json:"log"`          // Path to the log folder
	Compression bool             `json:"compression"`  // Enable/Disable compression
	Command     []string         `json:"command"`      // The command to run
	Notify      NotificationList `json:"notification"` // Notification systems list
}

type EmailNotification struct {
	ID       int              `json:"id"`
	Type     NotificationType `json:"type"` // "email"
	From     string           `json:"from"`
	To       []string         `json:"to"`
	Password string           `json:"password"`
	Server   string           `json:"server"`
	Port     int              `json:"port"`
	SSL      bool             `json:"ssl"`
}

func (e *EmailNotification) Kind() string { return "email" }

type WebhookNotification struct {
	ID          int               `json:"id"`
	Type        NotificationType  `json:"type"`         // "webhook"
	WebhookType Webhook           `json:"webhook_type"` // "discord"
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Events      []EventType       `json:"events"`
	Timeout     float64           `json:"timeout"`
	Headers     map[string]string `json:"headers"` // null -> nil is fine
	MaxRetries  int               `json:"max_retries"`
}

func (w *WebhookNotification) Kind() string { return "webhook" }

// GetObject returns the email object associated with the notification ty
func (t *NotificationType) GetObject() (Notification, error) {
	switch *t {
	case EmailNotify:
		return &EmailNotification{}, nil
	case WebhookNotify:
		return &WebhookNotification{}, nil
	default:
		return nil, fmt.Errorf("unknown notification type %q", *t)
	}
}

// UnmarshalJSON Helper to unmarshal polymorphic type Notification
// in the Job configuration JSON. By implementing this method
// the NotificationList type will be considered as implementing
// the Unmarshaler interface defined in the json package.
func (l *NotificationList) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*l = nil
		return nil
	}

	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	out := make([]Notification, 0, len(raw))
	for _, item := range raw {
		// Create a temporary struct to peek the type
		var peek struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(item, &peek); err != nil {
			return err
		}

		if peek.Type == "" {
			return fmt.Errorf("notification missing required field 'type'")
		}

		t := NotificationType(peek.Type)
		e, err := t.GetObject()
		if err != nil {
			return err
		}

		if err := json.Unmarshal(item, e); err != nil {
			return err
		}

		out = append(out, e)
	}

	*l = out
	return nil
}

// ToJson unmarshal the input string into the generic Type.
// If the type is not marshable, or json encodable then it
// returns an error.
func ToJson[T any](jsonStr string) (T, error) {
	var object T // Create the object of type T
	err := json.Unmarshal([]byte(jsonStr), &object)
	return object, err
}

func ToJsonWithObj[T any](dst *T, jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), dst)
	return err
}

// This struct identify a single job.
type Job struct {
	Id     int       // The unique identifier of the job
	Name   string    // The name of the job (unique as well)
	Status JobStatus // The status of the current job (enabled/disabled)
	Config JobConfig // Job configuration
}
