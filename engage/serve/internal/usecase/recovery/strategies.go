package recovery

// RecoveryStrategies returns ordered recovery actions for an error type (HexStrike parity).
func (h *Handler) RecoveryStrategies(t ErrorType) []Strategy {
	switch t {
	case TypeTimeout:
		return []Strategy{
			{ActionRetryWithBackoff, map[string]any{"initial_delay": 5, "max_delay": 60}, 3, 2.0, 0.7, 30},
			{ActionRetryWithReducedScope, map[string]any{"reduce_threads": true, "reduce_timeout": true}, 2, 1.0, 0.8, 45},
			{ActionSwitchToAlternativeTool, map[string]any{"use_fallback": true}, 1, 1.0, 0.6, 60},
		}
	case TypePermissionDenied:
		return []Strategy{
			{ActionEscalateToHuman, map[string]any{"message": "Permission denied - manual intervention required", "urgency": "high"}, 1, 1.0, 0.9, 300},
			{ActionSwitchToAlternativeTool, map[string]any{"use_sudo": true}, 1, 1.0, 0.3, 120},
		}
	case TypeNetworkUnreachable:
		return []Strategy{
			{ActionRetryWithBackoff, map[string]any{"initial_delay": 10, "max_delay": 120}, 3, 2.0, 0.6, 60},
			{ActionSwitchToAlternativeTool, map[string]any{"use_proxy": true}, 1, 1.0, 0.5, 90},
		}
	case TypeRateLimited:
		return []Strategy{
			{ActionRetryWithBackoff, map[string]any{"initial_delay": 30, "max_delay": 300}, 3, 2.0, 0.8, 120},
			{ActionAdjustParameters, map[string]any{"reduce_rate": true, "add_delay": true}, 2, 1.0, 0.7, 60},
		}
	case TypeToolNotFound:
		return []Strategy{
			{ActionSwitchToAlternativeTool, map[string]any{"use_fallback": true}, 1, 1.0, 0.9, 30},
			{ActionEscalateToHuman, map[string]any{"message": "Tool not found - installation required", "urgency": "medium"}, 1, 1.0, 0.9, 600},
		}
	case TypeInvalidParams:
		return []Strategy{
			{ActionAdjustParameters, map[string]any{"validate_params": true}, 2, 1.0, 0.8, 15},
			{ActionSwitchToAlternativeTool, map[string]any{"use_defaults": true}, 1, 1.0, 0.6, 30},
		}
	case TypeResourceExhausted:
		return []Strategy{
			{ActionRetryWithReducedScope, map[string]any{"reduce_threads": true, "reduce_memory": true}, 2, 1.0, 0.7, 60},
			{ActionRetryWithBackoff, map[string]any{"initial_delay": 60, "max_delay": 300}, 2, 2.0, 0.5, 180},
		}
	case TypeAuthenticationFailed:
		return []Strategy{
			{ActionEscalateToHuman, map[string]any{"message": "Authentication failed - check credentials", "urgency": "high"}, 1, 1.0, 0.9, 300},
			{ActionSwitchToAlternativeTool, map[string]any{"use_anon": true}, 1, 1.0, 0.2, 60},
		}
	case TypeTargetUnreachable:
		return []Strategy{
			{ActionRetryWithBackoff, map[string]any{"initial_delay": 15, "max_delay": 180}, 3, 2.0, 0.5, 90},
			{ActionGracefulDegradation, map[string]any{"skip_target": true}, 1, 1.0, 0.8, 30},
		}
	case TypeParsing:
		return []Strategy{
			{ActionAdjustParameters, map[string]any{"fix_format": true}, 2, 1.0, 0.7, 20},
			{ActionSwitchToAlternativeTool, map[string]any{"use_simple_format": true}, 1, 1.0, 0.6, 30},
		}
	default:
		return []Strategy{
			{ActionRetryWithBackoff, map[string]any{"initial_delay": 5, "max_delay": 60}, 2, 2.0, 0.4, 60},
			{ActionEscalateToHuman, map[string]any{"message": "Unknown error encountered", "urgency": "medium"}, 1, 1.0, 0.9, 300},
		}
	}
}

// PrimaryAction returns the first recovery action for an error type.
func (h *Handler) PrimaryAction(t ErrorType) RecoveryAction {
	strategies := h.RecoveryStrategies(t)
	if len(strategies) == 0 {
		return ActionEscalateToHuman
	}
	return strategies[0].Action
}
