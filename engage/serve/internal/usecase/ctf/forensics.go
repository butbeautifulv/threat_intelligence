package ctf

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	toolsuc "github.com/butbeautifulv/veil/engage/serve/internal/usecase/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// ForensicsOptions configures forensics analysis.
type ForensicsOptions struct {
	AnalysisType      string
	ExtractHidden     bool
	CheckSteganography bool
	FilesDir          string
}

// ForensicsAnalysis is the forensics-analyzer response shape.
type ForensicsAnalysis struct {
	FilePath            string         `json:"file_path"`
	AnalysisType        string         `json:"analysis_type"`
	FileInfo            map[string]any `json:"file_info"`
	Metadata            map[string]any `json:"metadata"`
	HiddenData          []any          `json:"hidden_data"`
	SteganographyResults []any         `json:"steganography_results"`
	RecommendedTools    []string       `json:"recommended_tools"`
	NextSteps           []string       `json:"next_steps"`
}

// AnalyzeForensics runs heuristic and optional tool-based forensics analysis.
func AnalyzeForensics(ctx context.Context, subject string, filePath string, opts ForensicsOptions, runner *toolsuc.Runner, reg *tools.Registry) ForensicsAnalysis {
	out := ForensicsAnalysis{
		FilePath:     filePath,
		AnalysisType: opts.AnalysisType,
		FileInfo:     map[string]any{},
		Metadata:     map[string]any{},
		HiddenData:   []any{},
		SteganographyResults: []any{},
		RecommendedTools: []string{"exiftool", "binwalk", "strings"},
		NextSteps:    []string{},
	}
	if filePath == "" {
		return out
	}
	resolved := resolveFilePath(filePath, opts.FilesDir)
	out.FilePath = resolved

	if ft, err := runCmd(ctx, "file", resolved); err == nil && ft != "" {
		out.FileInfo["type"] = strings.TrimSpace(ft)
		lower := strings.ToLower(ft)
		if strings.Contains(lower, "image") {
			out.RecommendedTools = append(out.RecommendedTools, "exiftool", "steghide", "zsteg")
			out.NextSteps = append(out.NextSteps, "Extract EXIF metadata", "Check for steganographic content")
		} else if strings.Contains(lower, "zip") || strings.Contains(lower, "archive") {
			out.RecommendedTools = append(out.RecommendedTools, "binwalk")
			out.NextSteps = append(out.NextSteps, "Extract archive contents")
		}
	}

	if runner != nil && reg != nil {
		for _, toolName := range []string{"exiftool_extract", "binwalk_analyze", "strings_extract"} {
			if _, err := reg.MustGet(toolName); err != nil {
				continue
			}
			res := runner.Run(ctx, subject, toolName, contract.ToolRunRequest{Target: resolved})
			out.HiddenData = append(out.HiddenData, map[string]any{
				"tool": toolName, "success": res.Success, "stdout": truncate(res.Output, 4000),
			})
		}
	}

	out.RecommendedTools = dedupStrings(out.RecommendedTools)
	return out
}

func resolveFilePath(path, filesDir string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if filesDir != "" {
		return filepath.Join(filesDir, path)
	}
	return path
}

func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// FileExists checks whether path is readable.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
