package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/butbeautifulv/veil/engage/serve/internal/audit"
	"github.com/butbeautifulv/veil/engage/serve/internal/domain/tool"
	"github.com/butbeautifulv/veil/engage/serve/internal/runner"
	"github.com/butbeautifulv/veil/engage/serve/internal/security"
	"github.com/butbeautifulv/veil/engage/serve/internal/tools"
	"github.com/butbeautifulv/veil/engage/serve/internal/telemetry"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/cache"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/recovery"
	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/engage/contract"
)

// Runner executes catalog tools.
type Runner struct {
	Registry *tools.Registry
	Exec     *runner.Executor
	Audit    *audit.Logger
	Auth     *auth.Stack
	Cache    *cache.Store
	Recovery    *recovery.Handler
	TargetGuard security.TargetGuardMode
}

func (r *Runner) recovery() *recovery.Handler {
	if r.Recovery != nil {
		return r.Recovery
	}
	return recovery.Default()
}

func (r *Runner) List() []tool.Spec {
	return r.Registry.List()
}

func (r *Runner) Run(ctx context.Context, subject string, name string, req contract.ToolRunRequest) contract.ToolRunResponse {
	res := r.runOnce(ctx, subject, name, req)
	if res.Success {
		return res
	}
	return r.recoverRun(ctx, subject, name, req, res, 0)
}

func (r *Runner) recoverRun(ctx context.Context, subject, name string, req contract.ToolRunRequest, last contract.ToolRunResponse, attempt int) contract.ToolRunResponse {
	h := r.recovery()
	errType := h.Classify(last.Error + " " + last.Output)
	if !h.Recoverable(errType) || attempt >= h.MaxRetries(errType) {
		return last
	}
	alt := h.SuggestAlternative(name, errType)
	if alt != "" && attempt == 0 {
		altName := tools.ResolveCatalogName(alt+"_scan", r.Registry)
		if spec, ok := r.Registry.Get(altName); ok && spec.Enabled {
			alt = altName
		} else if _, ok := r.Registry.Get(alt); !ok {
			alt = tools.ResolveCatalogName(alt, r.Registry)
		}
		retry := r.runOnce(ctx, subject, alt, req)
		if retry.Success {
			retry.Tool = name
			retry.Output = "[recovery: " + alt + "] " + retry.Output
			return retry
		}
		last = retry
	}
	delay := h.BackoffDelay(attempt + 1)
	if delay > 0 {
		select {
		case <-ctx.Done():
			return last
		case <-time.After(delay):
		}
	}
	params := h.AdjustParams(name, errType, mergeParametersFromReq(name, req, r.Registry))
	retry := r.runOnce(ctx, subject, name, contract.ToolRunRequest{
		Target:     req.Target,
		Parameters: params,
	})
	if retry.Success {
		retry.Output = "[recovery: retry] " + retry.Output
		return retry
	}
	return r.recoverRun(ctx, subject, name, req, retry, attempt+1)
}

func mergeParametersFromReq(name string, req contract.ToolRunRequest, reg *tools.Registry) map[string]string {
	spec, err := reg.MustGet(name)
	if err != nil {
		return req.Parameters
	}
	return mergeParameters(spec, req)
}

func (r *Runner) emitAudit(subject, name, target, jobID string, ok bool, errMsg string) {
	if r.Audit != nil && strings.TrimSpace(name) != "" {
		r.Audit.ToolRun(subject, name, target, jobID, ok, errMsg)
	}
}

func (r *Runner) runOnce(ctx context.Context, subject string, name string, req contract.ToolRunRequest) contract.ToolRunResponse {
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
	guardMode := r.TargetGuard
	if guardMode == "" {
		guardMode = security.TargetGuardOff
	}
	if blocked, reason := security.CheckTarget(req.Target); blocked {
		switch guardMode {
		case security.TargetGuardBlock:
			msg := "target blocked by ENGAGE_TARGET_GUARD: " + reason
			r.emitAudit(subject, name, req.Target, fmt.Sprintf("%s-block-%d", name, time.Now().UnixNano()), false, msg)
			return contract.ToolRunResponse{Success: false, Tool: name, Error: msg}
		case security.TargetGuardWarn:
			// allow but audit note is optional — emit as failed=false with stderr prefix in output only on block
		}
	}
	spec, err := r.Registry.MustGet(name)
	if err != nil {
		r.emitAudit(subject, name, req.Target, fmt.Sprintf("%s-err-%d", name, time.Now().UnixNano()), false, err.Error())
		return contract.ToolRunResponse{Success: false, Tool: name, Error: err.Error()}
	}
	bin, err := runner.LookupBinary(spec.Binary)
	if err != nil {
		errMsg := err.Error()
		r.emitAudit(subject, name, req.Target, fmt.Sprintf("%s-err-%d", name, time.Now().UnixNano()), false, errMsg)
		return contract.ToolRunResponse{Success: false, Tool: name, Error: errMsg}
	}
	params := mergeParameters(spec, req)
	args := runner.BuildArgs(spec.ArgsTemplate, req.Target, req.AdditionalArgs, params)
	cacheKey := fmt.Sprintf("tool:%s:%s:%v", name, req.Target, params)
	if req.Parameters != nil && req.Parameters["use_cache"] == "false" {
		cacheKey = ""
	}
	if cacheKey != "" && r.Cache != nil {
		if out, ok := r.Cache.Get(cacheKey); ok {
			return contract.ToolRunResponse{Success: true, Tool: name, Output: out, JobID: "cached"}
		}
	}
	timeout := time.Duration(spec.TimeoutSec) * time.Second
	track := &runner.TrackInfo{Tool: name, Target: req.Target}
	res := r.Exec.Run(ctx, bin, args, timeout, track)
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
	if ok && cacheKey != "" && r.Cache != nil {
		r.Cache.Set(cacheKey, out)
	}
	r.emitAudit(subject, name, req.Target, jobID, ok, errMsg)
	telemetry.RecordToolRun(name, ok)
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
		for _, alias := range []string{"url", "domain"} {
			if specHasParam(spec, alias) && out[alias] == "" {
				out[alias] = req.Target
			}
		}
	}
	return out
}

func specHasParam(spec tool.Spec, name string) bool {
	for _, p := range spec.Parameters {
		if p.Name == name {
			return true
		}
	}
	return false
}
