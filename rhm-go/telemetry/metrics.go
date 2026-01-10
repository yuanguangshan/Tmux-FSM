package telemetry

import (
	"fmt"
	"rhm-go/core/analysis"
	"rhm-go/core/history"
	"rhm-go/core/solver"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	SolveDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rhm_solve_duration_seconds",
		Help:    "Time taken to resolve conflicts",
		Buckets: []float64{0.01, 0.1, 0.5, 1, 5},
	}, []string{"complexity", "result"})

	ConflictCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rhm_conflict_count",
		Help: "Number of conflicts detected",
	}, []string{"severity"})

	MemoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "rhm_memory_usage_bytes",
		Help: "Current memory consumption",
	})
)

func RegisterMetrics() {
	prometheus.MustRegister(SolveDuration)
	prometheus.MustRegister(ConflictCount)
	prometheus.MustRegister(MemoryUsage)
}

func InstrumentSolver(originalSolver func(*history.HistoryDAG, history.NodeID, history.NodeID) solver.ResolutionPlan) func(*history.HistoryDAG, history.NodeID, history.NodeID) solver.ResolutionPlan {
	return func(dag *history.HistoryDAG, tipA, tipB history.NodeID) solver.ResolutionPlan {
		start := time.Now()
		complexity := len(dag.Nodes)

		result := originalSolver(dag, tipA, tipB)

		duration := time.Since(start).Seconds()
		resultLabel := "failure"
		if result.Resolved {
			resultLabel = "success"
		}

		SolveDuration.WithLabelValues(fmt.Sprint(complexity), resultLabel).Observe(duration)

		// 内存采样
		go func() {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			MemoryUsage.Set(float64(m.Alloc))
		}()

		return result
	}
}

// RecordConflictRecord 记录冲突监控
func RecordConflictRecord(c analysis.Conflict) {
	severity := "low"
	sev := analysis.ConflictSeverity(c)
	if sev >= 100 {
		severity = "high"
	} else if sev >= 80 {
		severity = "medium"
	}

	ConflictCount.WithLabelValues(severity).Inc()
}
