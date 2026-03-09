package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
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

type LogConfig struct {
	Path            string `json:"path"`             // Path to the log folder
	MaxSpareFiles   int    `json:"max_spare_files"`  // Maximum number of spare files
	RetentionWindow int    `json:"retention_window"` // Maximum windows in days before removing some logs
}

type JobConfig struct {
	Name        string           `json:"name"`         // The name of the job
	Log         LogConfig        `json:"log"`          // Path to the log folder
	Compression bool             `json:"compression"`  // Enable/Disable compression
	Command     []string         `json:"command"`      // The command to run
	Notify      NotificationList `json:"notification"` // Notification systems list
	Schedule    string           `json:"schedule"`     // String schedule specification
}

type EmailNotification struct {
	Id       int              `json:"id"`
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
	Id          int               `json:"id"`
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

// This struct identify a single job.
type Job struct {
	Id     int       // The unique identifier of the job
	Name   string    // The name of the job (unique as well)
	Status JobStatus // The status of the current job (enabled/disabled)
	Config JobConfig // Job configuration
}

func (j Job) Duration() time.Duration {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	schedule, _ := parser.Parse(j.Config.Schedule)
	next := schedule.Next(time.Now())
	return time.Until(next)
}

type RunStatus int

const (
	RunStatusWaiting RunStatus = iota
	RunStatusRunning
	RunStatusCompleted
)

func (r RunStatus) String() string {
	switch r {
	case RunStatusWaiting:
		return "waiting"
	case RunStatusRunning:
		return "running"
	case RunStatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

type Run struct {
	Id      string    // The Id of the run
	Status  RunStatus // Current status of the run
	JobName string    // Name of the job of the current run
	DryRun  bool      // If the run is a dry-run
	OneShot bool      // If this run is one shot or not
}

type JobRun struct {
	Run

	Job *Job // Job of the current run
}

func (r JobRun) Duration() time.Duration {
	return r.Job.Duration()
}
