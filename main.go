package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/previousnext/prometheus-healthz/internal/prometheus"
)

const (
	// EnvarPort for overriding the default port.
	EnvarPort = "PROMETHEUS_HEALTHZ_PORT"
	// EnvarPath for overriding the default path.
	EnvarPath = "PROMETHEUS_HEALTHZ_PATH"
	// EnvarQuery for overriding the default query.
	EnvarQuery = "PROMETHEUS_HEALTHZ_QUERY"
	// EnvarURI for overriding the default URI.
	EnvarURI = "PROMETHEUS_HEALTHZ_URI"
	// EnvarUsername for overriding the default username.
	EnvarUsername = "PROMETHEUS_HEALTHZ_USERNAME"
	// EnvarPassword for overriding the default password.
	EnvarPassword = "PROMETHEUS_HEALTHZ_PASSWORD"
)

var (
	cliPort     = kingpin.Flag("port", "Port which to serv requests").Default(":80").Envar(EnvarPort).String()
	cliPath     = kingpin.Flag("path", "Path which to serv requests").Default("/healthz").Envar(EnvarPath).String()
	cliQuery    = kingpin.Flag("query", "Username used for basic authentication").Default("type=healthz").Envar(EnvarQuery).String()
	cliURI      = kingpin.Flag("uri", "Promtheus endpoint").Default("http://127.0.0.1:9090").Envar(EnvarURI).String()
	cliUsername = kingpin.Flag("username", "Username used for basic authentication").Required().Envar(EnvarUsername).String()
	cliPassword = kingpin.Flag("password", "Password used for basic authentication").Required().Envar(EnvarPassword).String()
)

func main() {
	kingpin.Parse()

	e := echo.New()

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == *cliUsername && password == *cliPassword {
			return true, nil
		}

		return false, nil
	}))

	e.GET(*cliPath, func(c echo.Context) error {
		client, err := prometheus.New(*cliURI)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		rules, err := client.Rules()
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		filtered, err := getHealthzRules(*cliQuery, rules)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		if len(filtered) > 0 {
			return c.JSON(http.StatusInternalServerError, Response{State: StateUnhealthy, Rules: filtered})
		}

		return c.JSON(http.StatusOK, Response{State: StateHealthy})
	})

	e.Logger.Fatal(e.Start(*cliPort))
}

// Helper function to return a list of rules which are "firing" and labelled "healthz".
func getHealthzRules(query string, resp prometheus.RulesResponse) ([]string, error) {
	var rules []string

	labels, err := getLabels(query)
	if err != nil {
		return rules, errors.Wrap(err, "failed to get labels from query")
	}

	for _, group := range resp.Data.Groups {
		for _, rule := range group.Rules {
			for labelKey, labelValue := range labels {
				if !hasLabel(labelKey, labelValue, rule) {
					continue
				}
			}

			if !isFiring(rule) {
				continue
			}

			rules = append(rules, rule.Name)
		}
	}

	return rules, nil
}

// Helper function to check if the "cluster healthz" label is applied to a rule.
func hasLabel(label, value string, rule prometheus.Rule) bool {
	if val, ok := rule.Labels[label]; ok {
		if val == value {
			return true
		}
	}

	return false
}

// Helper function to check if a rule is "firing" an alert.
func isFiring(rule prometheus.Rule) bool {
	for _, alert := range rule.Alerts {
		if alert.State == prometheus.StateFiring {
			return true
		}
	}

	return false
}

// Helper function to extract a query into a key/value label pair.
func getLabels(query string) (map[string]string, error) {
	labels := make(map[string]string)

	list := strings.Split(query, ",")

	for _, item := range list {
		q := strings.Split(item, "=")

		if len(q) == 2 {
			var (
				key = q[0]
				val = q[1]
			)

			labels[key] = val
		}
	}

	return labels, nil
}
