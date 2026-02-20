package service

import (
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strconv"

	apirequests "github.com/lmriccardo/backer/deamon/internal/api/http/v1/requests"
	"github.com/lmriccardo/backer/deamon/internal/domain"
	"github.com/lmriccardo/backer/deamon/internal/platform/constants"
	"github.com/lmriccardo/backer/deamon/internal/platform/utils"
)

var TIMEOUT_RE *regexp.Regexp = regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:e(\d+))?(s|ms|us)$`)

func fillSmtpServer(dst *domain.EmailNotification, smtp *apirequests.SMTPConfig) error {
	domain_name := utils.GetDomainFromEmail(dst.From)
	smtp_provider, ok := constants.SMTP_PROVIDERS[domain_name]
	if !ok && smtp == nil {
		return NewConfigurationError(
			"EmailNotification",
			"uknown smtp provider and SMTP configuration is not set",
		)
	}

	// If the SMTP provider is known, then we directly
	// take the values from the map
	if ok {
		dst.Server = smtp_provider.Hostname
		dst.Port = int(smtp_provider.Port)
		dst.SSL = smtp_provider.Ssl
		return nil
	}

	// Otherwise, we use the one provided by the user
	dst.Server = smtp.Server
	dst.Port = int(smtp.Port)
	dst.SSL = smtp.Ssl

	return nil
}

func timeoutStrToFloat(timeout string) (float64, error) {
	matches := TIMEOUT_RE.FindStringSubmatch(timeout)
	if matches == nil {
		return -1, NewConfigurationError(
			"WebhookNotification:Timeout",
			fmt.Sprintf("invalid timeout string: %v", timeout),
		)
	}

	intPart := matches[1] // Take the unit part
	decPart := matches[2] // Take the decimal part
	expPart := matches[3] // Take the exponential part
	unit := matches[4]    // Take the measure unit (seconds, milli, micro)

	result, _ := strconv.ParseFloat(intPart, 64)

	if decPart != "" {
		result, _ = strconv.ParseFloat(intPart+"."+decPart, 64)
	}

	if expPart != "" {
		exp, _ := strconv.Atoi(expPart)
		result *= math.Pow10(exp)

	}

	switch unit {
	case "ms":
		result /= 1000
	case "us":
		result /= 1_000_000
	}

	return result, nil
}

func createRsyncCommand(job *apirequests.CreateJobRequest) []string {
	command := []string{"rsync"}

	flags := "-aHAX"
	if job.Rsync.Options.Verbose {
		flags = "-avvHAX"
	}

	command = append(command, flags)
	command = append(command, fmt.Sprintf("--password-file=%s", job.Remote.Password))
	command = append(command, "--delete")
	command = append(command, fmt.Sprintf("--delete-%s", job.Rsync.Options.Delete))

	if job.Rsync.Options.ShowProgress {
		command = append(command, "--info=progress2")
	}

	command = append(command, "--prune-empty-dirs")

	// When putting includes and excludes, I would like to prioritize
	// excludes, for this reason I will put includes first and excludes
	// after. Using this pattern, if any excludes appears also in the
	// include list it will be automatically excluded.
	for _, includes := range job.Rsync.Includes {
		command = append(command, fmt.Sprintf("--include=%s", includes))
	}

	for _, excludes := range job.Rsync.Excludes {
		command = append(command, fmt.Sprintf("--exclude=%s", excludes))
	}

	command = append(command, fmt.Sprintf("--exclude-from=%s", job.Rsync.ExcludeFrom))
	command = append(command, "--numeric-ids")

	if job.Rsync.Options.ItemizeChanges {
		command = append(command, "--itemize-changes")
	}

	if !job.Rsync.Options.KeepDevices {
		command = append(command, "--no-devices")
	}

	if !job.Rsync.Options.KeepSpecials {
		command = append(command, "--no-specials")
	}

	// Add all the sources before the remote destination
	command = append(command, job.Rsync.Sources...)

	// Add the host, port, user, module and folder
	rsync_host := fmt.Sprintf("rsync://%s@%s:%d/%s/%s/",
		job.Remote.User, job.Remote.Host, job.Remote.Port,
		job.Remote.Dest.Module, job.Remote.Dest.Folder,
	)

	command = append(command, rsync_host)

	return command
}

func createJob(job *apirequests.CreateJobRequest) (*domain.Job, error) {
	job_config := domain.JobConfig{Name: job.Name}

	// 1. Configure log setting
	log_folder, err := utils.BackerLogHome()
	if err != nil {
		return nil, err
	}

	log_folder = filepath.Join(log_folder, job.Name)
	job_config.Log = domain.LogConfig{
		Path: log_folder, MaxSpareFiles: job.LogRetention.MaxSpareFiles,
		RetentionWindow: job.LogRetention.Window,
	}

	// 2. Create the rsync command
	job_config.Command = createRsyncCommand(job)

	// 3. Configure notification system. First check for email
	// notification system if present in the job input request.
	curr_notif_id := 1
	if job.Notify.Email != nil {
		// First format generic informations about the email
		// notification system.
		email_notif := &domain.EmailNotification{
			Id: curr_notif_id, Type: domain.EmailNotify,
			From: job.Notify.Email.From, To: job.Notify.Email.To,
			Password: job.Notify.Email.Password,
		}

		// Then we need to setup the SMTP server
		if err := fillSmtpServer(email_notif, job.Notify.Email.Smtp); err != nil {
			return nil, err
		}

		job_config.Notify = append(job_config.Notify, email_notif)
		curr_notif_id++
	}

	// Then check for any webhook notification
	for _, wbhook_ntfy := range job.Notify.Webhooks {
		timeout_v, err := timeoutStrToFloat(wbhook_ntfy.Timeout)
		if err != nil {
			return nil, err
		}

		job_config.Notify = append(job_config.Notify,
			&domain.WebhookNotification{
				Id:          curr_notif_id,
				Type:        domain.WebhookNotify,
				WebhookType: domain.WebhookDiscord,
				Name:        wbhook_ntfy.Name,
				URL:         wbhook_ntfy.URL,
				Events:      wbhook_ntfy.Events,
				Timeout:     timeout_v,
				MaxRetries:  wbhook_ntfy.MaxRetries,
				Headers:     wbhook_ntfy.Headers,
			},
		)

		curr_notif_id++
	}

	return &domain.Job{
		Id:     -1,
		Name:   job.Name,
		Status: domain.JobStatusEnabled,
		Config: job_config,
	}, nil
}
