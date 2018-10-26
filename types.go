package main

type State string

const (
	StateHealthy   State = "healthy"
	StateUnhealthy State = "unhealthy"
)

type Response struct {
	State State    `json:"state"`
	Rules []string `json:"rules"`
}
