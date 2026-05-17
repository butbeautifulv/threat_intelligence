package tooldispatch

// workflowBridgeBinaries matches scripts/engage/generate-tools-na-matrix.py WORKFLOW_BINARIES.
var workflowBridgeBinaries = map[string]struct{}{
	"api": {}, "bugbounty": {}, "ai": {}, "get": {}, "http": {}, "create": {}, "execute": {},
	"generate": {}, "list": {}, "kube": {}, "browser": {}, "autorecon": {}, "comprehensive": {},
	"advanced": {}, "analyze": {}, "checkov": {}, "clair": {}, "cloudmapper": {}, "checksec": {},
	"clear": {},
}

// IsBridgeWorkflowBinary reports catalog tools classified as bridge_api via workflow placeholder binaries.
func IsBridgeWorkflowBinary(binary string) bool {
	_, ok := workflowBridgeBinaries[binary]
	return ok
}
