package service

import "fmt"

type JobDuplicateError struct {
	Name string // The name of the duplicate job
}

func (j *JobDuplicateError) Error() string {
	return fmt.Sprintf("duplicated job name %v", j.Name)
}

type InvalidJobNameError struct {
	Name string // The job name
}

func (i *InvalidJobNameError) Error() string {
	return fmt.Sprintf("invalid job name %v", i.Name)
}
