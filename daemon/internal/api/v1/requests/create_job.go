package requests

import (
	"fmt"

	"github.com/lmriccardo/backer/deamon/internal/domain"
)

// JobCreateRequest is the body structure of the JSON data
// for the POST http request to endpoint /jobs/create for
// creating a new Job.
type CreateJobRequest struct {
	Name         string          `json:"name" example:"full_backup" binding:"required"` // The name of the target/job
	Remote       RemoteConfig    `json:"remote"                     binding:"required"` // The remote configuration
	Rsync        RsyncConfig     `json:"rsync"                      binding:"required"` // Rsync configuration options
	Schedule     ScheduleConfig  `json:"schedule"                   binding:"required"` // Job Schedule configuration
	Notify       NotificationCfg `json:"notification"`                                  // Notification systems configuration
	LogRetention LogRetentionCfg `json:"log_retention"`                                 // Log retention configuration
}

// Configuration for the remote server
// NOTE: tags live here to avoid duplicating API request structs.
type RemoteConfig struct {
	Host     string        `json:"host"     example:"host.domain"   binding:"required"`        // Server Hostname/IPv4 address
	Port     uint16        `json:"port"     example:"1234"          binding:"min=1,max=65535"` // Remote server port
	User     string        `json:"user"     example:"admin"         binding:"required"`        // The username to connect with
	Password string        `json:"password" example:"password-file" binding:"required"`        // Password file for non-interactive
	Dest     RemoteDestCfg `json:"dest"                             binding:"required"`        // Remote destination module and folder
}

func (r *RemoteConfig) ApplyDefault() {
	if r.Port == 0 {
		r.Port = 873
	}
}

// Configuration for the remote destination of the backup
// in the form <module>/<folder> as required by rsync
type RemoteDestCfg struct {
	Module string `json:"module" example:"module_name" binding:"required"` // The module name
	Folder string `json:"folder" example:"dest_folder" binding:"required"` // The destination folder inside the module
}

// RSync configuration for the backup plan
type RsyncConfig struct {
	ExcludeOutputFolder string       `json:"exclude_output_folder"`      // Folder where the expanded excludes are saved into.
	ExcludeFrom         string       `json:"exclude_from"`               // Base file with a bunch of paths to exclude
	Excludes            []string     `json:"excludes"`                   // Additional spare exclude paths
	Includes            []string     `json:"includes"`                   // Folders to include ( removes from excludes if present )
	Sources             []string     `json:"sources" binding:"required"` // Sources to copy into the destination
	Options             RsyncOptions `json:"options"`                    // Rsync option configuration
}

type DeleteType string

const (
	DeleteInvalid  DeleteType = ""
	DeleteBegin    DeleteType = "begin"
	DeleteAfter    DeleteType = "after"
	DeleteDuring   DeleteType = "during"
	DeleteDelay    DeleteType = "delay"
	DeleteExcluded DeleteType = "excluded"
)

// Rsync options for command formatting
type RsyncOptions struct {
	Verbose        bool       `json:"verbose"`                                      // Enable/Disable verbosity
	ShowProgress   bool       `json:"show_progress"`                                // Shows the progress in the log output
	Compress       bool       `json:"compress"`                                     // Enable or Disable compression before trasmitting
	Delete         DeleteType `json:"delete" example:"after" binding:"delete_mode"` // Delete mode one of [ before, after, during, delay and excluded ]
	ItemizeChanges bool       `json:"itemize_changes"`                              // Output a change-summary for all updates
	KeepSpecials   bool       `json:"keep_specials"`                                // Keep specials files likes UNIX sockets or FIFOs.
	KeepDevices    bool       `json:"keep_devices"`                                 // Keep devices file (`/dev/null`, `/dev/sda`, ...)
}

func (r *RsyncOptions) ApplyDefault() {
	if r.Delete == DeleteInvalid {
		r.Delete = DeleteAfter
	}
}

// Job Schedule configuration in Cronjob format
type ScheduleConfig struct {
	Weekday string `json:"weekday" example:"0" binding:"required"` // Range 0-7 (Sun = 0 or 7)
	Month   string `json:"month"   example:"*" binding:"required"` // Range 1-12
	Day     string `json:"day"     example:"*" binding:"required"` // Day of the month with Range 1-31
	Hour    string `json:"hour"    example:"6" binding:"required"` // Range 0-23
	Minute  string `json:"minute"  example:"0" binding:"required"` // Range 0-59
}

func (s *ScheduleConfig) String() string {
	return fmt.Sprintf("%s %s %s %s %s",
		s.Minute, s.Hour, s.Day, s.Month, s.Weekday,
	)
}

// Custom SMTP server configuration
type SMTPConfig struct {
	Server string `json:"server" binding:"required"`                 // The SMTP server host
	Port   uint16 `json:"port"   binding:"required,min=1,max=65535"` // The SMTP Server port
	Ssl    bool   `json:"ssl"    binding:"required"`                 // Tell if the SMTP server has SSL enabled or not
}

// Notification system configuration
type NotificationCfg struct {
	Email    *EmailNotification    `json:"email"`    // Email notification system configuration
	Webhooks []WebhookNotification `json:"webhooks"` // Webhooks notification system list
}

// Email notification system. If the SMTP server can be inferred from the
// email domain than the SMTP section is not required, otherwise it is
// required and if not present an error occurrs.
type EmailNotification struct {
	From     string      `json:"from"     binding:"required,email"`            // The email of the sender
	To       []string    `json:"to"       binding:"required,min=1,dive,email"` // A list of recipients for the notification email
	Password string      `json:"password" binding:"required"`                  // The account password used to authentication to the SMTP server
	Smtp     *SMTPConfig `json:"smtp"`                                         // Optional SMTP server
}

// Webhooks notifications system
type WebhookNotification struct {
	Name       string             `json:"name"        binding:"required"`                  // The name of the service (there can be duplicates)
	Type       domain.Webhook     `json:"type"        binding:"required,webhook_type"`     // The type of the webhook service endpoint
	URL        string             `json:"url"         binding:"required,http_url"`         // The endpoint URL of the service
	Events     []domain.EventType `json:"events"      binding:"omitempty,dive,event_type"` // Subscribed Events: failure, success
	Timeout    string             `json:"timeout"     binding:"timeout_type"`              // Timeout for receiving the response after sending the notification.
	MaxRetries int                `json:"max_retries" binding:"min=0"`                     // Maximum number of retries
	Headers    map[string]string  `json:"headers"`                                         // Additional headers used for the requests
}

func (w *WebhookNotification) ApplyDefault() {
	if len(w.Events) == 0 {
		w.Events = []domain.EventType{domain.EventSuccess, domain.EventFailure}
	}

	if w.Timeout == "" {
		w.Timeout = "0s"
	}
}

// Log retention configuration
type LogRetentionCfg struct {
	MaxSpareFiles int `json:"max_spare_files" example:"10" binding:"min=0"` // Maximum number of spare log files
	Window        int `json:"retention_window" example:"7" binding:"min=0"` // Log retention window in days
}

func (l *LogRetentionCfg) ApplyDefault() {
	if l.MaxSpareFiles < 1 {
		l.MaxSpareFiles = 10
	}

	if l.Window < 1 {
		l.Window = 7
	}
}
