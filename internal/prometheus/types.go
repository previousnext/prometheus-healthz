package prometheus

type RulesResponse struct {
	Data RulesData `json:"data"`
}

type RulesData struct {
	Groups []RulesGroup `json:"groups"`
}

type RulesGroup struct {
	Name  string `json:"name"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Alerts []Alert           "json:alerts"
}

type State string

const (
	StateInactive State = "inactive"
	StateFiring   State = "firing"
	StatePending  State = "pending"
)

type Alert struct {
	Labels map[string]string `json:"labels"`
	State  State             `json:"state"`
}
