package metrics

import (
	"context"
	"time"

	"github.com/Proton-105/himera-bot/internal/state"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	botCommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_commands_total",
			Help: "Total number of bot commands received labeled by command and status",
		},
		[]string{"command", "status"},
	)
	commandDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "command_duration_seconds",
			Help:    "Duration of bot commands in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"command"},
	)
	stateTransitionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "state_transitions_total",
			Help: "Total number of state transitions",
		},
		[]string{"from", "to"},
	)
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors split by type and severity",
		},
		[]string{"type", "severity"},
	)
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users",
			Help: "Current number of active users",
		},
	)
	usersByState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "users_by_state",
			Help: "Number of users per state",
		},
		[]string{"state"},
	)
)

var trackedStates = []state.State{
	state.StateIdle,
	state.StateBuyingSearch,
	state.StateBuyingAmount,
	state.StateBuyingConfirm,
	state.StateError,
}

func init() {
	state.RegisterTransitionRecorder(RecordStateTransition)
}

// RecordCommand increments command counters and records duration.
func RecordCommand(command, status string, duration time.Duration) {
	if command == "" {
		command = "unknown"
	}
	if status == "" {
		status = "unknown"
	}

	botCommandsTotal.WithLabelValues(command, status).Inc()
	commandDurationSeconds.WithLabelValues(command).Observe(duration.Seconds())
}

// RecordStateTransition tracks FSM transitions.
func RecordStateTransition(from, to string) {
	if from == "" {
		from = "unknown"
	}
	if to == "" {
		to = "unknown"
	}

	stateTransitionsTotal.WithLabelValues(from, to).Inc()
}

// RecordError increments error counters with metadata.
func RecordError(errType, severity string) {
	if errType == "" {
		errType = "unknown"
	}
	if severity == "" {
		severity = "unknown"
	}

	errorsTotal.WithLabelValues(errType, severity).Inc()
}

// SetActiveUsers updates the gauge for current active users.
func SetActiveUsers(count int) {
	activeUsers.Set(float64(count))
}

// SetUsersByState updates the gauge for the given state.
func SetUsersByState(state string, count int) {
	if state == "" {
		state = "unknown"
	}

	usersByState.WithLabelValues(state).Set(float64(count))
}

// StateCollector periodically gathers FSM state counts and emits gauge metrics.
type StateCollector struct {
	fsm state.StateMachine
}

// NewStateCollector builds a metrics collector bound to the provided FSM.
func NewStateCollector(fsm state.StateMachine) *StateCollector {
	return &StateCollector{fsm: fsm}
}

// Run polls the FSM every 10 seconds, updating active user gauges until ctx is cancelled.
func (c *StateCollector) Run(ctx context.Context) {
	if c == nil || c.fsm == nil {
		return
	}
	if ctx == nil {
		ctx = context.Background()
	}

	for {
		_ = c.collect(ctx)

		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
		}
	}
}

func (c *StateCollector) collect(ctx context.Context) error {
	states, err := c.fsm.GetAllStates(ctx)
	if err != nil {
		return err
	}

	SetActiveUsers(len(states))

	stateCounts := make(map[string]int, len(states))
	for _, st := range states {
		label := "unknown"
		if st != nil && st.CurrentState != "" {
			label = string(st.CurrentState)
		}
		stateCounts[label]++
	}

	usersByState.Reset()

	for _, tracked := range trackedStates {
		label := string(tracked)
		SetUsersByState(label, stateCounts[label])
		delete(stateCounts, label)
	}

	for label, count := range stateCounts {
		SetUsersByState(label, count)
	}

	return nil
}
