package tooldispatch

import (
	"context"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
)

type ctfBridgeHandler func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error)

var ctfBridgeHandlers = map[string]ctfBridgeHandler{
	"ctf_create_challenge_workflow": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		ch := ctf.ChallengeFromBody(args)
		ch.Name = firstNonEmpty(ch.Name, target, "challenge")
		ch.Target = firstNonEmpty(ch.Target, target)
		return d.CTF.CreateChallengeWorkflow(ch)
	},
	"ctf_auto_solve_challenge": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		ch := ctf.ChallengeFromBody(args)
		ch.Name = firstNonEmpty(ch.Name, target, "challenge")
		ch.Target = firstNonEmpty(ch.Target, target)
		exec := argString(args, "execute_tools", "true") != "false"
		return d.CTF.AutoSolve(ctx, subject, ch, exec, argInt(args, "max_steps", 8))
	},
	"ctf_suggest_tools": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		_ = ctx
		_ = subject
		desc := argString(args, "description", "")
		if desc == "" {
			return nil, dispatchToolError("description required")
		}
		return d.CTF.SuggestTools(desc, argString(args, "category", "misc"), target), nil
	},
	"ctf_team_strategy": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		_ = ctx
		_ = subject
		_ = target
		_ = args
		return map[string]any{
			"success": true,
			"note":    "use HTTP POST /api/ctf/team-strategy with challenges[] and team_skills",
		}, nil
	},
	"ctf_cryptography_solver": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		_ = target
		text := argString(args, "cipher_text", "")
		if text == "" {
			return nil, dispatchToolError("cipher_text required")
		}
		return d.CTF.AnalyzeCrypto(text, argString(args, "cipher_type", "unknown"),
			argString(args, "key_hint", ""), argString(args, "known_plaintext", ""),
			argString(args, "additional_info", "")), nil
	},
	"ctf_forensics_analyzer": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		_ = target
		path := argString(args, "file_path", "")
		if path == "" {
			return nil, dispatchToolError("file_path required")
		}
		return d.CTF.AnalyzeForensics(ctx, subject, path, ctf.ForensicsOptions{}), nil
	},
	"ctf_binary_analyzer": func(ctx context.Context, d *Dispatcher, subject, target string, args map[string]any) (any, error) {
		_ = target
		path := argString(args, "binary_path", "")
		if path == "" {
			return nil, dispatchToolError("binary_path required")
		}
		return d.CTF.AnalyzeBinary(ctx, subject, path, ctf.BinaryOptions{}), nil
	},
}

func (d *Dispatcher) callCTFBridge(ctx context.Context, name, subject, target string, args map[string]any) (any, error) {
	if d.CTF == nil {
		return nil, dispatchToolError("ctf service not configured")
	}
	if h, ok := ctfBridgeHandlers[name]; ok {
		return h(ctx, d, subject, target, args)
	}
	return nil, dispatchNotFound("unknown ctf tool: %s", name)
}
