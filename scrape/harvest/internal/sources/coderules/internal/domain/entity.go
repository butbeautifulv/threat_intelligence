package domain

// RuleFile identifies a Semgrep/CodeQL rules file from GitHub.
type RuleFile struct {
	Repo   string
	Path   string
	Format string
}
