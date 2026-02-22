package v1

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/lmriccardo/backer/deamon/internal/domain"
	"github.com/robfig/cron/v3"
)

func validateEnum[T comparable](fl validator.FieldLevel, value ...T) bool {
	s, ok := fl.Field().Interface().(T)
	if !ok {
		return false
	}

	return slices.Contains(value, s)
}

func deleteModeValidator(fl validator.FieldLevel) bool {
	return validateEnum(fl, DeleteBegin, DeleteAfter,
		DeleteDelay, DeleteDuring, DeleteExcluded,
	)
}

func webhookTypeValidator(fl validator.FieldLevel) bool {
	return validateEnum(fl, domain.WebhookDiscord)
}

func eventTypeValidator(fl validator.FieldLevel) bool {
	return validateEnum(fl, domain.EventFailure, domain.EventSuccess)
}

func timeoutValidator(fl validator.FieldLevel) bool {
	s, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	pattern := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:e(\d+))?(s|ms|us)$`)
	return pattern.MatchString(s)
}

func scheduleValidator(sl validator.StructLevel) {
	s := sl.Current().Interface().(ScheduleConfig)

	spec := fmt.Sprintf("%s %s %s %s %s",
		s.Minute, s.Hour, s.Day, s.Month, s.Weekday,
	)

	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	if _, err := parser.Parse(spec); err != nil {
		sl.ReportError(s, "ScheduleConfig", "schedule", "cron", err.Error())
		return
	}

	// Optional strict rule:
	// prevent both Day and Weekday being specific (classic cron ambiguity)
	if s.Day != "*" && s.Weekday != "*" {
		sl.ReportError(s, "ScheduleConfig", "schedule", "cron_conflict",
			"cannot specify both day-of-month and weekday")
	}
}

var VALIDATORS = map[string]validator.Func{
	"delete_mode": deleteModeValidator, "webhook_type": webhookTypeValidator,
	"event_type": eventTypeValidator, "timeout_type": timeoutValidator,
}

// RegisterValidators registers custom validators for Gin
func RegisterValidators() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		// gin is not using go-playground validator (rare)
		return nil
	}

	// Registers field level validators
	for tag, vfn := range VALIDATORS {
		if err := v.RegisterValidation(tag, vfn); err != nil {
			return err
		}
	}

	// Registers struct level validators
	v.RegisterStructValidation(scheduleValidator, ScheduleConfig{})

	return nil
}

func FormatValidatorError(err error) ([]error, bool) {
	if verrs, ok := err.(validator.ValidationErrors); ok {
		out := make([]error, 0, len(verrs))
		for _, e := range verrs {
			out = append(out, fmt.Errorf(
				"Validation failed for field %s (tag %s, param %s)",
				e.Field(), e.Tag(), e.Param(),
			))
		}
		return out, true
	}
	return nil, false
}

func ValidateRequest[T any](req T) (bool, []error) {
	if err := binding.Validator.ValidateStruct(req); err != nil {
		if verr, ok := FormatValidatorError(err); ok {
			return false, verr
		}
		return false, []error{err}
	}
	return true, nil
}
