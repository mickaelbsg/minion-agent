package heartbeat

import (
	"time"

	"minion/internal/agentinfo"
)

var processStartedAt = time.Now().UTC()

// Status is a lightweight liveness snapshot intended for control-plane polling.
type Status struct {
	Status               string  `json:"status"`
	AgentID              string  `json:"agent_id"`
	Hostname             string  `json:"hostname"`
	Version              string  `json:"version"`
	SystemUptimeSeconds  float64 `json:"system_uptime_seconds"`
	ProcessUptimeSeconds float64 `json:"process_uptime_seconds"`
	ProcessStartedAt     string  `json:"process_started_at"`
	ObservedAt           string  `json:"observed_at"`
}

// Get returns the current heartbeat using the same identity as /api/v1/agent.
func Get() Status {
	return At(time.Now().UTC())
}

// At builds a heartbeat for a supplied observation time and supports deterministic tests.
func At(now time.Time) Status {
	if now.Before(processStartedAt) {
		now = processStartedAt
	}
	info := agentinfo.Get()
	return Status{
		Status:               "online",
		AgentID:              info.AgentID,
		Hostname:             info.Hostname,
		Version:              info.Version,
		SystemUptimeSeconds:  info.UptimeSeconds,
		ProcessUptimeSeconds: now.Sub(processStartedAt).Seconds(),
		ProcessStartedAt:     processStartedAt.Format(time.RFC3339),
		ObservedAt:           now.Format(time.RFC3339),
	}
}
