package domain

import (
	engagedomainjob "github.com/butbeautifulv/veil/pkg/engage/domain/job"
	engagedomainreport "github.com/butbeautifulv/veil/pkg/engage/domain/report"
	engagedomaintarget "github.com/butbeautifulv/veil/pkg/engage/domain/target"
	engagedomaintool "github.com/butbeautifulv/veil/pkg/engage/domain/tool"
)

// Engage contour type aliases (SOT remains pkg/engage/domain/*).
type (
	Target   = engagedomaintarget.Target
	Finding  = engagedomainreport.Finding
	Severity = engagedomainreport.Severity
	Job      = engagedomainjob.Job
	Status   = engagedomainjob.Status
	Spec     = engagedomaintool.Spec
	Param    = engagedomaintool.Param
)

// Severity re-exports.
const (
	SeverityInfo     = engagedomainreport.SeverityInfo
	SeverityLow      = engagedomainreport.SeverityLow
	SeverityMedium   = engagedomainreport.SeverityMedium
	SeverityHigh     = engagedomainreport.SeverityHigh
	SeverityCritical = engagedomainreport.SeverityCritical
)

// Job status re-exports.
const (
	StatusPending   = engagedomainjob.StatusPending
	StatusRunning   = engagedomainjob.StatusRunning
	StatusDone      = engagedomainjob.StatusDone
	StatusFailed    = engagedomainjob.StatusFailed
	StatusCancelled = engagedomainjob.StatusCancelled
)
