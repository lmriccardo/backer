package service

import "fmt"

type DuplicateJobNameError struct {
	Name string // The name of the duplicate job
}

func NewDuplicateJobNameError(name string) *DuplicateJobNameError {
	return &DuplicateJobNameError{Name: name}
}

func (j *DuplicateJobNameError) Error() string {
	return fmt.Sprintf("duplicated job name %v", j.Name)
}

type InvalidJobNameError struct {
	Name string // The job name
}

func NewInvalidJobNameError(name string) *InvalidJobNameError {
	return &InvalidJobNameError{Name: name}
}

func (i *InvalidJobNameError) Error() string {
	return fmt.Sprintf("invalid job name %v", i.Name)
}

type ConfigurationError struct {
	Level   string // The configuration level at which there is the error
	Message string // The error message for that part
}

func NewConfigurationError(lvl string, msg string) *ConfigurationError {
	return &ConfigurationError{lvl, msg}
}

func (c *ConfigurationError) Error() string {
	return fmt.Sprintf("configuration error ( %v ): %v", c.Level, c.Message)
}

type DatabaseError struct {
	Message string // The error describing message
}

func NewDatabaseError(msg string) *DatabaseError {
	return &DatabaseError{msg}
}

func (d *DatabaseError) Error() string {
	return fmt.Sprintf("database error: %v", d.Message)
}
