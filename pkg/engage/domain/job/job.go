package job

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusDone      Status = "done"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// Job represents an async long-running tool scan.
type Job struct {
	ID         string            `json:"id"`
	ToolName   string            `json:"tool_name"`
	Target     string            `json:"target"`
	Subject    string            `json:"subject,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
	Status     Status            `json:"status"`
	Output     string            `json:"output,omitempty"`
	Error      string            `json:"error,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
