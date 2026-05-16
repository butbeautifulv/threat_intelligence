package ctf

import (
	"context"
	"strings"

	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// BinaryOptions configures binary analysis.
type BinaryOptions struct {
	AnalysisDepth    string
	CheckProtections bool
	FindGadgets      bool
	FilesDir         string
}

// BinaryAnalysis is the binary-analyzer response shape.
type BinaryAnalysis struct {
	BinaryPath         string         `json:"binary_path"`
	AnalysisDepth      string         `json:"analysis_depth"`
	FileInfo           map[string]any `json:"file_info"`
	SecurityProtections map[string]any `json:"security_protections"`
	InterestingStrings []string       `json:"interesting_strings"`
	RecommendedTools   []string       `json:"recommended_tools"`
	ExploitationHints  []string       `json:"exploitation_hints"`
	ToolResults        []any          `json:"tool_results,omitempty"`
}

// AnalyzeBinary runs checksec/strings-style analysis via catalog tools when available.
func AnalyzeBinary(ctx context.Context, subject, binaryPath string, opts BinaryOptions, runner *toolsuc.Runner, reg *tools.Registry) BinaryAnalysis {
	out := BinaryAnalysis{
		BinaryPath:         binaryPath,
		AnalysisDepth:      opts.AnalysisDepth,
		FileInfo:           map[string]any{},
		SecurityProtections: map[string]any{},
		InterestingStrings: []string{},
		RecommendedTools:   []string{"checksec", "strings", "file", "ghidra", "radare2"},
		ExploitationHints:  []string{},
	}
	if binaryPath == "" {
		return out
	}
	resolved := resolveFilePath(binaryPath, opts.FilesDir)
	out.BinaryPath = resolved

	if ft, err := runCmd(ctx, "file", resolved); err == nil {
		out.FileInfo["type"] = strings.TrimSpace(ft)
	}

	if runner != nil && reg != nil {
		for _, toolName := range []string{"checksec_analyze", "strings_extract"} {
			if _, err := reg.MustGet(toolName); err != nil {
				continue
			}
			res := runner.Run(ctx, subject, toolName, contract.ToolRunRequest{Target: resolved})
			entry := map[string]any{"tool": toolName, "success": res.Success, "stdout": truncate(res.Output, 8000)}
			out.ToolResults = append(out.ToolResults, entry)
			if res.Success && res.Output != "" {
				if toolName == "checksec_analyze" {
					out.SecurityProtections["checksec"] = res.Output
				}
				for _, line := range strings.Split(res.Output, "\n") {
					lower := strings.ToLower(line)
					if strings.Contains(lower, "flag") || strings.Contains(lower, "password") || strings.Contains(lower, "secret") {
						out.InterestingStrings = append(out.InterestingStrings, strings.TrimSpace(line))
					}
				}
			}
		}
	}

	if len(out.InterestingStrings) > 20 {
		out.InterestingStrings = out.InterestingStrings[:20]
	}
	if opts.FindGadgets {
		out.ExploitationHints = append(out.ExploitationHints, "Use ropper/ROPgadget for ROP chain building")
	}
	if opts.CheckProtections {
		out.ExploitationHints = append(out.ExploitationHints, "Review NX/PIE/Canary from checksec output")
	}
	out.RecommendedTools = dedupStrings(out.RecommendedTools)
	return out
}
