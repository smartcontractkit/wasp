package wasp

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AlertChecker is checking alerts according to dashboardUUID and requirements labels
type AlertChecker struct {
	URL                  string
	APIKey               string
	RequirementLabelName string
	T                    *testing.T
	l                    zerolog.Logger
	client               *resty.Client
}

type Alert struct {
	Annotations struct {
		DashboardUID string `json:"__dashboardUid__"`
		OrgID        string `json:"__orgId__"`
		PanelID      string `json:"__panelId__"`
		Description  string `json:"description"`
		RunbookURL   string `json:"runbook_url"`
		Summary      string `json:"summary"`
	} `json:"annotations"`
	EndsAt      time.Time `json:"endsAt"`
	Fingerprint string    `json:"fingerprint"`
	Receivers   []struct {
		Active       interface{} `json:"active"`
		Integrations interface{} `json:"integrations"`
		Name         string      `json:"name"`
	} `json:"receivers"`
	StartsAt time.Time `json:"startsAt"`
	Status   struct {
		InhibitedBy []interface{} `json:"inhibitedBy"`
		SilencedBy  []interface{} `json:"silencedBy"`
		State       string        `json:"state"`
	} `json:"status"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	GeneratorURL string            `json:"generatorURL"`
	Labels       map[string]string `json:"labels"`
}

// AlertGroupsResponse is response body for "api/alertmanager/grafana/api/v2/alerts/groups"
type AlertGroupsResponse struct {
	Alerts []Alert `json:"alerts"`
	Labels struct {
		Alertname     string `json:"alertname"`
		GrafanaFolder string `json:"grafana_folder"`
	} `json:"labels"`
	Receiver struct {
		Active       interface{} `json:"active"`
		Integrations interface{} `json:"integrations"`
		Name         string      `json:"name"`
	} `json:"receiver"`
}

func NewAlertChecker(t *testing.T, requirenemtLabelName string) *AlertChecker {
	url := os.Getenv("GRAFANA_URL")
	if url == "" {
		panic(fmt.Errorf("GRAFANA_URL env var must be defined"))
	}
	apiKey := os.Getenv("GRAFANA_TOKEN")
	if apiKey == "" {
		panic(fmt.Errorf("GRAFANA_TOKEN env var must be defined"))
	}
	return &AlertChecker{
		URL:                  url,
		APIKey:               apiKey,
		RequirementLabelName: requirenemtLabelName,
		T:                    t,
		client:               resty.New(),
		l:                    GetLogger(t, "AlertChecker"),
	}
}

// AnyAlerts check if any alerts with dashboardUUID have been raised
func (m *AlertChecker) AnyAlerts(dashboardUUID, requirementLabelValue string) error {
	alerts := make([]Alert, 0)
	raised := false
	defer func() {
		if m.T != nil && raised {
			m.T.Fail()
		}
	}()
	var result []AlertGroupsResponse
	_, err := m.client.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", m.APIKey)).
		SetResult(&result).
		Get(fmt.Sprintf("%s/api/alertmanager/grafana/api/v2/alerts/groups", m.URL))
	if err != nil {
		return fmt.Errorf("failed to get alert groups: %s", err)
	}
	for _, a := range result {
		for _, aa := range a.Alerts {
			if aa.Annotations.DashboardUID == dashboardUUID && aa.Labels[m.RequirementLabelName] == requirementLabelValue {
				log.Warn().
					Str("Summary", aa.Annotations.Summary).
					Str("Description", aa.Annotations.Description).
					Str("URL", aa.GeneratorURL).
					Interface("Labels", aa.Labels).
					Time("StartsAt", aa.StartsAt).
					Time("UpdatedAt", aa.UpdatedAt).
					Str("State", aa.Status.State).
					Msg("Alert fired")
				alerts = append(alerts, aa)
				raised = true
			}
		}
	}
	return nil
}
