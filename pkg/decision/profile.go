package decision

import "strings"

// TargetProfile is the Go port of HexStrike TargetProfile used by the decision engine.
type TargetProfile struct {
	Target             string
	TargetType         string
	Technologies       []string
	CMS                string
	IPAddresses        []string
	OpenPorts          int
	AttackSurfaceScore float64
	ConfidenceScore    float64
	RiskLevel          string
}

var attackSurfaceTypeBase = map[string]float64{
	"web":     7.0,
	"api":     6.0,
	"ip":      8.0,
	"cloud":   5.0,
	"binary":  4.0,
	"unknown": 3.0,
}

// BuildTargetProfile assembles a profile from probe output (HexStrike analyze_target parity).
func BuildTargetProfile(target, targetType string, technologies []string, cms string, ips []string, openPorts int) TargetProfile {
	p := TargetProfile{
		Target:       target,
		TargetType:   targetType,
		Technologies: technologies,
		CMS:          cms,
		IPAddresses:  ips,
		OpenPorts:    openPorts,
	}
	p.AttackSurfaceScore = calculateAttackSurface(p)
	p.RiskLevel = riskFromAttackSurface(p.AttackSurfaceScore)
	p.ConfidenceScore = calculateConfidence(p)
	return p
}

func calculateAttackSurface(p TargetProfile) float64 {
	score := attackSurfaceTypeBase[p.TargetType]
	if score == 0 {
		score = attackSurfaceTypeBase["unknown"]
	}
	score += float64(len(p.Technologies)) * 0.5
	score += float64(p.OpenPorts) * 0.3
	if p.CMS != "" {
		score += 1.5
	}
	if score > 10 {
		return 10
	}
	return score
}

func riskFromAttackSurface(score float64) string {
	switch {
	case score >= 8:
		return "critical"
	case score >= 6:
		return "high"
	case score >= 4:
		return "medium"
	case score >= 2:
		return "low"
	default:
		return "minimal"
	}
}

func calculateConfidence(p TargetProfile) float64 {
	conf := 0.5
	if len(p.IPAddresses) > 0 {
		conf += 0.1
	}
	if len(p.Technologies) > 0 && !sliceContainsFold(p.Technologies, "unknown") {
		conf += 0.2
	}
	if p.CMS != "" {
		conf += 0.1
	}
	if p.TargetType != "unknown" {
		conf += 0.1
	}
	return min(conf, 1.0)
}

func sliceContainsFold(ss []string, want string) bool {
	want = strings.ToLower(want)
	for _, s := range ss {
		if strings.EqualFold(s, want) {
			return true
		}
	}
	return false
}

func profileHasTech(p TargetProfile, labels ...string) bool {
	for _, want := range labels {
		want = strings.ToLower(want)
		for _, t := range p.Technologies {
			if strings.Contains(strings.ToLower(t), want) {
				return true
			}
		}
		if strings.EqualFold(p.CMS, want) {
			return true
		}
	}
	return false
}
