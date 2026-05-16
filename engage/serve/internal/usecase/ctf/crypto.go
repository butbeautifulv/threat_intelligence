package ctf

import (
	"regexp"
	"strings"
)

// CryptoAnalysis is the cryptography-solver response shape.
type CryptoAnalysis struct {
	CipherText         string   `json:"cipher_text"`
	CipherType         string   `json:"cipher_type"`
	AnalysisResults    []string `json:"analysis_results"`
	PotentialSolutions []string `json:"potential_solutions"`
	RecommendedTools   []string `json:"recommended_tools"`
	NextSteps          []string `json:"next_steps"`
}

// AnalyzeCrypto applies heuristic cipher analysis (no LLM).
func AnalyzeCrypto(cipherText, cipherType, keyHint, knownPlaintext, additionalInfo string) CryptoAnalysis {
	out := CryptoAnalysis{
		CipherText:       cipherText,
		CipherType:       cipherType,
		AnalysisResults:  []string{},
		RecommendedTools: []string{},
		NextSteps:        []string{},
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(cipherText, " ", ""), "\n", "")
	lower := strings.ToLower(additionalInfo + " " + keyHint)

	if cipherType == "unknown" || cipherType == "" {
		if matched, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, clean); matched && len(clean) > 8 {
			out.AnalysisResults = append(out.AnalysisResults, "Possible hexadecimal encoding")
			out.RecommendedTools = append(out.RecommendedTools, "hex")
		}
		if matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]+=*$`, strings.TrimSpace(cipherText)); matched {
			out.AnalysisResults = append(out.AnalysisResults, "Possible Base64 encoding")
			out.RecommendedTools = append(out.RecommendedTools, "base64")
		}
	}

	hashLens := map[int]string{32: "MD5", 40: "SHA1", 64: "SHA256", 128: "SHA512"}
	if t, ok := hashLens[len(clean)]; ok {
		if matched, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, clean); matched {
			out.AnalysisResults = append(out.AnalysisResults, "Possible "+t+" hash")
			out.RecommendedTools = append(out.RecommendedTools, "hashcat", "john")
		}
	}

	if cipherType == "caesar" || strings.Contains(lower, "rot") {
		out.RecommendedTools = append(out.RecommendedTools, "rot13")
		out.NextSteps = append(out.NextSteps, "Try all ROT values (1-25)")
	}
	if cipherType == "rsa" || strings.Contains(lower, "rsa") {
		out.RecommendedTools = append(out.RecommendedTools, "rsatool")
		out.NextSteps = append(out.NextSteps, "Check if modulus can be factored")
	}
	if cipherType == "vigenere" || strings.Contains(lower, "vigenere") {
		out.NextSteps = append(out.NextSteps, "Perform Kasiski examination for key length")
	}
	if knownPlaintext != "" {
		out.NextSteps = append(out.NextSteps, "Apply known-plaintext attack with provided plaintext")
	}

	out.RecommendedTools = dedupStrings(out.RecommendedTools)
	return out
}

func dedupStrings(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
