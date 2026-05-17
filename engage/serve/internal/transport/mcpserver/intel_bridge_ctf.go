package mcpserver

import (
	"context"

	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
)

type ctfBridgeHandler func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error)

var ctfBridgeHandlers = map[string]ctfBridgeHandler{
	"ctf_create_challenge_workflow": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		ch := ctf.ChallengeFromBody(args)
		ch.Name = firstNonEmpty(ch.Name, target, "challenge")
		ch.Target = firstNonEmpty(ch.Target, target)
		out, err := s.ctf.CreateChallengeWorkflow(ch)
		if err != nil {
			return nil, rpcErrf(codeToolError, "%v", err)
		}
		return toolJSONResult(out)
	},
	"ctf_auto_solve_challenge": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		ch := ctf.ChallengeFromBody(args)
		ch.Name = firstNonEmpty(ch.Name, target, "challenge")
		ch.Target = firstNonEmpty(ch.Target, target)
		exec := argString(args, "execute_tools", "true") != "false"
		out, err := s.ctf.AutoSolve(ctx, subject, ch, exec, argInt(args, "max_steps", 8))
		if err != nil {
			return nil, rpcErrf(codeToolError, "%v", err)
		}
		return toolJSONResult(out)
	},
	"ctf_suggest_tools": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = ctx
		_ = subject
		desc := argString(args, "description", "")
		if desc == "" {
			return nil, rpcErrf(codeToolError, "description required")
		}
		return toolJSONResult(s.ctf.SuggestTools(desc, argString(args, "category", "misc"), target))
	},
	"ctf_team_strategy": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = ctx
		_ = subject
		_ = target
		_ = args
		return toolJSONResult(map[string]any{
			"success": true,
			"note":    "use HTTP POST /api/ctf/team-strategy with challenges[] and team_skills",
		})
	},
	"ctf_cryptography_solver": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = target
		text := argString(args, "cipher_text", "")
		if text == "" {
			return nil, rpcErrf(codeToolError, "cipher_text required")
		}
		return toolJSONResult(s.ctf.AnalyzeCrypto(text, argString(args, "cipher_type", "unknown"),
			argString(args, "key_hint", ""), argString(args, "known_plaintext", ""),
			argString(args, "additional_info", "")))
	},
	"ctf_forensics_analyzer": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = target
		path := argString(args, "file_path", "")
		if path == "" {
			return nil, rpcErrf(codeToolError, "file_path required")
		}
		return toolJSONResult(s.ctf.AnalyzeForensics(ctx, subject, path, ctf.ForensicsOptions{}))
	},
	"ctf_binary_analyzer": func(ctx context.Context, s *Server, subject, target string, args map[string]any) (map[string]any, error) {
		_ = target
		path := argString(args, "binary_path", "")
		if path == "" {
			return nil, rpcErrf(codeToolError, "binary_path required")
		}
		return toolJSONResult(s.ctf.AnalyzeBinary(ctx, subject, path, ctf.BinaryOptions{}))
	},
}

func (s *Server) callCTFBridge(ctx context.Context, name, subject, target string, args map[string]any) (map[string]any, error) {
	if s.ctf == nil {
		return nil, rpcErrf(codeToolError, "ctf service not configured")
	}
	if h, ok := ctfBridgeHandlers[name]; ok {
		return h(ctx, s, subject, target, args)
	}
	return nil, rpcErrf(codeMethodNotFound, "unknown ctf tool: %s", name)
}
