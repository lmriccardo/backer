package requests

type RunJobRequest struct {
	Name string // Job name

	DryRun bool `json:"dry_run" binding:"required"` // Runs a dry-run or a true run
	Notify bool `json:"notify"  binding:"required"` // Enable/Disable notifications
	Log    bool `json:"log"     binding:"required"` // Enable/Disable logging
}
