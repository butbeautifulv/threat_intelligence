package domain

import (
	"testing"

	engagedomainjob "github.com/butbeautifulv/veil/pkg/engage/domain/job"
	engagedomainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	engagedomaintarget "github.com/butbeautifulv/veil/pkg/engage/domain/target"
	engagedomaintool "github.com/butbeautifulv/veil/pkg/engage/domain/tool"
)

func TestEngageAliasesCompile(t *testing.T) {
	var _ Target = engagedomaintarget.Target{}
	var _ Finding = engagedomainreport.Finding{}
	var _ Job = engagedomainjob.Job{}
	var _ Spec = engagedomaintool.Spec{}
	if SeverityHigh != engagedomainreport.SeverityHigh {
		t.Fatal("SeverityHigh alias mismatch")
	}
	if StatusRunning != engagedomainjob.StatusRunning {
		t.Fatal("StatusRunning alias mismatch")
	}
}
