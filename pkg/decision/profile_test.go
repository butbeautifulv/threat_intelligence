package decision

import "testing"

func TestBuildTargetProfile_attackSurfaceAndRisk(t *testing.T) {
	tests := []struct {
		name       string
		targetType string
		techs      []string
		cms        string
		ports      int
		wantRisk   string
		wantMax10  bool
	}{
		{
			name:       "low risk unknown baseline",
			targetType: "unknown",
			wantRisk:   "low",
		},
		{
			name:       "medium risk cloud",
			targetType: "cloud",
			wantRisk:   "medium",
		},
		{
			name:       "high risk api baseline",
			targetType: "api",
			wantRisk:   "high",
		},
		{
			name:       "critical web with cms",
			targetType: "web",
			cms:        "wordpress",
			wantRisk:   "critical",
		},
		{
			name:       "critical caps at 10",
			targetType: "ip",
			techs:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			ports:      20,
			cms:        "drupal",
			wantRisk:   "critical",
			wantMax10:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BuildTargetProfile("target", tt.targetType, tt.techs, tt.cms, []string{"10.0.0.1"}, tt.ports)
			if p.RiskLevel != tt.wantRisk {
				t.Fatalf("RiskLevel = %q, want %q (score %.2f)", p.RiskLevel, tt.wantRisk, p.AttackSurfaceScore)
			}
			if tt.wantMax10 && p.AttackSurfaceScore != 10 {
				t.Fatalf("AttackSurfaceScore = %v, want cap 10", p.AttackSurfaceScore)
			}
		})
	}
}

func TestBuildTargetProfile_confidence(t *testing.T) {
	tests := []struct {
		name      string
		targetTyp string
		techs     []string
		cms       string
		ips       []string
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "baseline half",
			targetTyp: "unknown",
			wantMin:   0.5,
			wantMax:   0.5,
		},
		{
			name:      "full confidence capped at 1",
			targetTyp: "web",
			techs:     []string{"php"},
			cms:       "wordpress",
			ips:       []string{"10.0.0.1"},
			wantMin:   1,
			wantMax:   1,
		},
		{
			name:      "unknown tech does not add tech bonus",
			targetTyp: "unknown",
			techs:     []string{"unknown"},
			wantMin:   0.5,
			wantMax:   0.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BuildTargetProfile("t", tt.targetTyp, tt.techs, tt.cms, tt.ips, 0)
			if p.ConfidenceScore < tt.wantMin || p.ConfidenceScore > tt.wantMax {
				t.Fatalf("ConfidenceScore = %v, want [%v,%v]", p.ConfidenceScore, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestProfileHasTech(t *testing.T) {
	p := BuildTargetProfile("https://x", "web", []string{"Apache Tomcat"}, "", nil, 0)
	if !profileHasTech(p, "tomcat") {
		t.Fatal("expected tomcat in technologies")
	}
	pCMS := BuildTargetProfile("https://x", "web", nil, "WordPress", nil, 0)
	if !profileHasTech(pCMS, "wordpress") {
		t.Fatal("expected cms match")
	}
	if profileHasTech(p, "ruby") {
		t.Fatal("unexpected tech match")
	}
}

func TestSliceContainsFold(t *testing.T) {
	if !sliceContainsFold([]string{"Unknown"}, "unknown") {
		t.Fatal("expected fold match")
	}
	if sliceContainsFold([]string{"nginx"}, "unknown") {
		t.Fatal("expected no match")
	}
}
