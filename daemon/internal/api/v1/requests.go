package v1

import "github.com/lmriccardo/backer/deamon/internal/api/v1/requests"

type CreateJobRequest = requests.CreateJobRequest
type RemoteConfig = requests.RemoteConfig
type RemoteDestCfg = requests.RemoteDestCfg
type RsyncConfig = requests.RsyncConfig
type DeleteType = requests.DeleteType

const (
	DeleteInvalid  = requests.DeleteInvalid
	DeleteBegin    = requests.DeleteBegin
	DeleteAfter    = requests.DeleteAfter
	DeleteDuring   = requests.DeleteDuring
	DeleteDelay    = requests.DeleteDelay
	DeleteExcluded = requests.DeleteExcluded
)

type RsyncOptions = requests.RsyncOptions
type ScheduleConfig = requests.ScheduleConfig
type SMTPConfig = requests.SMTPConfig
type NotificationCfg = requests.NotificationCfg
type EmailNotification = requests.EmailNotification
type WebhookNotification = requests.WebhookNotification
type LogRetentionCfg = requests.LogRetentionCfg
