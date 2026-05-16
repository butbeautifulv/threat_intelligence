package ctf

import "testing"

func TestAnalyzeCrypto_hexAndHash(t *testing.T) {
	out := AnalyzeCrypto("d41d8cd98f00b204e9800998ecf8427e", "unknown", "", "", "")
	if len(out.RecommendedTools) == 0 && len(out.AnalysisResults) == 0 {
		t.Fatal("expected some analysis")
	}
}

func TestExtractFlagCandidates(t *testing.T) {
	flags := extractFlagCandidates("here is flag{test123} end")
	if len(flags) != 1 || flags[0] != "flag{test123}" {
		t.Fatalf("%v", flags)
	}
}

func TestValidateFlagFormat(t *testing.T) {
	if !validateFlagFormat("flag{abc}") {
		t.Fatal("expected valid flag")
	}
}
