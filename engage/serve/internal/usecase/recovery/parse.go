package recovery

import "strings"

// ParseErrorType accepts HexStrike legacy names and engage aliases.
func ParseErrorType(s string) (ErrorType, bool) {
	switch strings.TrimSpace(strings.ToLower(s)) {
	case string(TypeTimeout):
		return TypeTimeout, true
	case string(TypePermissionDenied), "permission":
		return TypePermissionDenied, true
	case string(TypeNetworkUnreachable):
		return TypeNetworkUnreachable, true
	case string(TypeRateLimited), "rate_limit":
		return TypeRateLimited, true
	case string(TypeToolNotFound), "not_found":
		return TypeToolNotFound, true
	case string(TypeInvalidParams):
		return TypeInvalidParams, true
	case string(TypeResourceExhausted):
		return TypeResourceExhausted, true
	case string(TypeAuthenticationFailed):
		return TypeAuthenticationFailed, true
	case string(TypeTargetUnreachable):
		return TypeTargetUnreachable, true
	case string(TypeParsing):
		return TypeParsing, true
	case string(TypeUnknown):
		return TypeUnknown, true
	default:
		return "", false
	}
}
