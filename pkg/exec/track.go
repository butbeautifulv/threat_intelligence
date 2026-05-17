package exec

// TrackInfo identifies a tool run for process admin APIs.
type TrackInfo struct {
	Tool   string
	Target string
}

// ProcessTracker records subprocess lifecycle (implemented by usecase/process.Manager).
type ProcessTracker interface {
	Register(pid int, tool, target, command string)
	RegisterDocker(tool, target, command, session string) int
	UpdateProgress(pid int, progress float64, lastOutput string, bytesProcessed int64)
	Finish(pid int, status string)
}
