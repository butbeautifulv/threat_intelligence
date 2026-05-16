package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// Runner executes catalog tools.
type Runner struct {
	Registry *tools.Registry
	Exec     *runner.Executor
	Audit    *audit.Logger
	Auth     *auth.Stack
}

func (r *Runner) List() []tool.Spec {
	return r.Registry.List()
}

func (r *Runner) Run(ctx context.Context, subject string, name string, req contract.ToolRunRequest) contract.ToolRunResponse {
	if r.Auth != nil && r.Auth.Config.Enabled {
		sub := &auth.Subject{Sub: subject}
		if subject != "" {
			if s, ok := auth.SubjectFromContext(ctx); ok {
				sub = s
			}
		}
		if err := r.Auth.Enforcer.Enforce(sub, auth.PermEngageToolRun); err != nil {
			return contract.ToolRunResponse{Success: false, Tool: name, Error: "forbidden"}
		}
	}
	spec, err := r.Registry.MustGet(name)
	if err != nil {
		return contract.ToolRunResponse{Success: false, Tool: name, Error: err.Error()}
	}
	bin, err := runner.LookupBinary(spec.Binary)
	if err != nil {
		return contract.ToolRunResponse{Success: false, Tool: name, Error: err.Error()}
	}
	params := mergeParameters(spec, req)
	args := runner.BuildArgs(spec.ArgsTemplate, req.Target, req.AdditionalArgs, params)
	timeout := time.Duration(spec.TimeoutSec) * time.Second
	res := r.Exec.Run(ctx, bin, args, timeout)
	jobID := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	out := res.Stdout
	if res.Stderr != "" {
		out = strings.TrimSpace(out + "\n" + res.Stderr)
	}
	ok := res.Err == nil && res.ExitCode == 0
	errMsg := ""
	if res.Err != nil {
		errMsg = res.Err.Error()
	} else if res.ExitCode != 0 {
		errMsg = fmt.Sprintf("exit code %d", res.ExitCode)
	}
	if r.Audit != nil {
		r.Audit.ToolRun(subject, name, req.Target, jobID, ok, errMsg)
	}
	return contract.ToolRunResponse{
		Success:  ok,
		Tool:     name,
		Output:   out,
		Error:    errMsg,
		ExitCode: res.ExitCode,
		JobID:    jobID,
	}
}

func mergeParameters(spec tool.Spec, req contract.ToolRunRequest) map[string]string {
	out := spec.DefaultParameters()
	if req.Parameters != nil {
		for k, v := range req.Parameters {
			out[k] = v
		}
	}
	if req.Target != "" {
		out["target"] = req.Target
	}
	return out
}
