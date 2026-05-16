package telemetry

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	once sync.Once
	reg  *prometheus.Registry

	toolRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "engage_tool_runs_total", Help: "Tool executions by tool and status"},
		[]string{"tool", "status"},
	)
	jobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "engage_jobs_total", Help: "Jobs by terminal status"},
		[]string{"status"},
	)
	auditEventsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{Name: "engage_audit_events_total", Help: "Audit events appended"},
	)
	jobsPending = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "engage_jobs_pending", Help: "Pending jobs"},
	)
	cacheEntries = prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "engage_cache_entries", Help: "Cache entries"},
	)
)

func registry() *prometheus.Registry {
	once.Do(func() {
		reg = prometheus.NewRegistry()
		reg.MustRegister(toolRunsTotal, jobsTotal, auditEventsTotal, jobsPending, cacheEntries)
	})
	return reg
}

// Handler returns the Prometheus scrape handler.
func Handler() http.Handler {
	return promhttp.HandlerFor(registry(), promhttp.HandlerOpts{})
}

// RecordToolRun increments tool run counter.
func RecordToolRun(tool string, success bool) {
	status := "failed"
	if success {
		status = "success"
	}
	toolRunsTotal.WithLabelValues(tool, status).Inc()
}

// RecordJob increments job status counter.
func RecordJob(status string) {
	jobsTotal.WithLabelValues(status).Inc()
}

// RecordAuditEvent increments audit counter.
func RecordAuditEvent() {
	auditEventsTotal.Inc()
}

// SetJobsPending updates pending gauge.
func SetJobsPending(n int) {
	jobsPending.Set(float64(n))
}

// SetCacheEntries updates cache gauge.
func SetCacheEntries(n int) {
	cacheEntries.Set(float64(n))
}
