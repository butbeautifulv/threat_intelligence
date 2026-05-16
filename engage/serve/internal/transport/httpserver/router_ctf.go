package httpserver

import (
	"net/http"

	"github.com/butbeautifulv/veil/engage/serve/internal/components"
	"github.com/butbeautifulv/veil/engage/serve/internal/usecase/ctf"
)

func registerCTF(mux *http.ServeMux, c *components.APIComponents) {
	if c.CTF == nil {
		return
	}
	postJSON(mux, "POST /api/ctf/create-challenge-workflow", func(r *http.Request, body map[string]any) (any, int) {
		ch := ctf.ChallengeFromBody(body)
		out, err := c.CTF.CreateChallengeWorkflow(ch)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}, http.StatusBadRequest
		}
		return out, http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/auto-solve-challenge", func(r *http.Request, body map[string]any) (any, int) {
		ch := ctf.ChallengeFromBody(body)
		exec := body["execute_tools"] == nil || body["execute_tools"] == true
		maxSteps := toInt(body["max_steps"], 8)
		out, err := c.CTF.AutoSolve(r.Context(), subject(r), ch, exec, maxSteps)
		if err != nil {
			return map[string]any{"success": false, "error": err.Error()}, http.StatusBadRequest
		}
		return out, http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/suggest-tools", func(r *http.Request, body map[string]any) (any, int) {
		desc := toString(body["description"])
		if desc == "" {
			return map[string]any{"success": false, "error": "description is required"}, http.StatusBadRequest
		}
		return c.CTF.SuggestTools(desc, toString(body["category"]), toString(body["target"])), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/team-strategy", func(r *http.Request, body map[string]any) (any, int) {
		raw, _ := body["challenges"].([]any)
		if len(raw) == 0 {
			return map[string]any{"success": false, "error": "challenges data is required"}, http.StatusBadRequest
		}
		var challenges []ctf.Challenge
		for _, item := range raw {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			challenges = append(challenges, ctf.ChallengeFromBody(m))
		}
		skills := map[string][]string{}
		if ts, ok := body["team_skills"].(map[string]any); ok {
			for member, v := range ts {
				if arr, ok := v.([]any); ok {
					for _, s := range arr {
						if str, ok := s.(string); ok {
							skills[member] = append(skills[member], str)
						}
					}
				}
			}
		}
		return c.CTF.TeamStrategy(challenges, skills), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/cryptography-solver", func(r *http.Request, body map[string]any) (any, int) {
		text := toString(body["cipher_text"])
		if text == "" {
			return map[string]any{"success": false, "error": "cipher text is required"}, http.StatusBadRequest
		}
		return c.CTF.AnalyzeCrypto(text, toString(body["cipher_type"]), toString(body["key_hint"]),
			toString(body["known_plaintext"]), toString(body["additional_info"])), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/forensics-analyzer", func(r *http.Request, body map[string]any) (any, int) {
		path := toString(body["file_path"])
		if path == "" {
			return map[string]any{"success": false, "error": "file path is required"}, http.StatusBadRequest
		}
		return c.CTF.AnalyzeForensics(r.Context(), subject(r), path, ctf.ForensicsOptions{
			AnalysisType:       toString(body["analysis_type"]),
			ExtractHidden:      body["extract_hidden"] != false,
			CheckSteganography: body["check_steganography"] != false,
		}), http.StatusOK
	})
	postJSON(mux, "POST /api/ctf/binary-analyzer", func(r *http.Request, body map[string]any) (any, int) {
		path := toString(body["binary_path"])
		if path == "" {
			return map[string]any{"success": false, "error": "binary path is required"}, http.StatusBadRequest
		}
		return c.CTF.AnalyzeBinary(r.Context(), subject(r), path, ctf.BinaryOptions{
			AnalysisDepth:    toString(body["analysis_depth"]),
			CheckProtections: body["check_protections"] != false,
			FindGadgets:      body["find_gadgets"] != false,
		}), http.StatusOK
	})
}
