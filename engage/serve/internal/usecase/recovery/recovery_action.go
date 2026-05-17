package recovery

// RecoveryAction mirrors HexStrike RecoveryAction enum values.
type RecoveryAction string

const (
	ActionRetryWithBackoff          RecoveryAction = "retry_with_backoff"
	ActionRetryWithReducedScope     RecoveryAction = "retry_with_reduced_scope"
	ActionSwitchToAlternativeTool   RecoveryAction = "switch_to_alternative_tool"
	ActionAdjustParameters          RecoveryAction = "adjust_parameters"
	ActionEscalateToHuman           RecoveryAction = "escalate_to_human"
	ActionGracefulDegradation       RecoveryAction = "graceful_degradation"
	ActionAbortOperation            RecoveryAction = "abort_operation"
)

// Strategy describes a ranked recovery option (HexStrike RecoveryStrategy subset).
type Strategy struct {
	Action              RecoveryAction `json:"action"`
	Parameters          map[string]any `json:"parameters"`
	MaxAttempts         int            `json:"max_attempts"`
	BackoffMultiplier   float64        `json:"backoff_multiplier"`
	SuccessProbability  float64        `json:"success_probability"`
	EstimatedTime       int            `json:"estimated_time"`
}
