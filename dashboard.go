package wasp

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/K-Phoen/grabana"
	"github.com/K-Phoen/grabana/alert"
	"github.com/K-Phoen/grabana/dashboard"
	"github.com/K-Phoen/grabana/logs"
	"github.com/K-Phoen/grabana/row"
	"github.com/K-Phoen/grabana/stat"
	"github.com/K-Phoen/grabana/target/prometheus"
	"github.com/K-Phoen/grabana/timeseries"
	"github.com/K-Phoen/grabana/timeseries/axis"
	"github.com/K-Phoen/grabana/variable/query"
)

const (
	DefaultStatTextSize       = 12
	DefaultStatValueSize      = 20
	DefaultAlertEvaluateEvery = "10s"
	DefaultAlertFor           = "10s"
	DefaultDashboardUUID      = "wasp"

	DefaultRequirementLabelKey = "requirement_name"
)

type WaspAlert struct {
	Name                 string
	AlertType            string
	TestName             string
	GenName              string
	RequirementGroupName string
	AlertIf              alert.ConditionEvaluator
	CustomAlert          timeseries.Option
}

// Dashboard is a Wasp dashboard
type Dashboard struct{}

// NewDashboard creates new dashboard instance
func NewDashboard() *Dashboard {
	return &Dashboard{}
}

// Deploy deploys this dashboard to some Grafana folder
func (m *Dashboard) Deploy(reqs []WaspAlert) (*grabana.Dashboard, error) {
	dsn := os.Getenv("DATA_SOURCE_NAME")
	if dsn == "" {
		return nil, fmt.Errorf("DATA_SOURCE_NAME must be provided")
	}
	dbf := os.Getenv("DASHBOARD_FOLDER")
	if dbf == "" {
		return nil, fmt.Errorf("DASHBOARD_FOLDER must be provided")
	}
	grafanaURL := os.Getenv("GRAFANA_URL")
	if grafanaURL == "" {
		return nil, fmt.Errorf("GRAFANA_URL must be provided")
	}
	grafanaToken := os.Getenv("GRAFANA_TOKEN")
	if grafanaToken == "" {
		return nil, fmt.Errorf("GRAFANA_TOKEN must be provided")
	}
	ctx := context.Background()
	d, err := m.Dashboard(dsn, reqs)
	if err != nil {
		return nil, fmt.Errorf("failed to build dashboard: %s", err)
	}
	client := grabana.NewClient(&http.Client{}, grafanaURL, grabana.WithAPIToken(grafanaToken))
	fo, err := client.FindOrCreateFolder(ctx, dbf)
	if err != nil {
		fmt.Printf("Could not find or create folder: %s\n", err)
		os.Exit(1)
	}
	return client.UpsertDashboard(ctx, fo, d)
}

const (
	AlertTypeQuantile99 = "quantile_99"
	AlertTypeErrors     = "errors"
)

func LokiAlertParams(queryType, testName, genName string) string {
	switch queryType {
	case AlertTypeQuantile99:
		return fmt.Sprintf(`
avg(quantile_over_time(0.99, {go_test_name="%s", test_data_type=~"responses", gen_name="%s"}
| json
| unwrap duration [10s]) / 1e6)`, testName, genName)
	case AlertTypeErrors:
		return fmt.Sprintf(`
max_over_time({go_test_name="%s", test_data_type=~"stats", gen_name="%s"}
| json
| unwrap failed [10s]) by (go_test_name, gen_name)`, testName, genName)
	default:
		return ""
	}
}

// defaultStatWidget creates default Stat widget
func defaultStatWidget(name, datasourceName, target, legend string) row.Option {
	return row.WithStat(
		name,
		stat.Transparent(),
		stat.DataSource(datasourceName),
		stat.Text(stat.TextValueAndName),
		stat.Orientation(stat.OrientationHorizontal),
		stat.TitleFontSize(DefaultStatTextSize),
		stat.ValueFontSize(DefaultStatValueSize),
		stat.Span(2),
		stat.WithPrometheusTarget(target, prometheus.Legend(legend)),
	)
}

// defaultLastValueAlertWidget creates default last value alert
func defaultLastValueAlertWidget(a WaspAlert) timeseries.Option {
	if a.CustomAlert != nil {
		return a.CustomAlert
	}
	return timeseries.Alert(
		a.Name,
		alert.For(DefaultAlertFor),
		alert.OnExecutionError(alert.ErrorKO),
		alert.Description(a.Name),
		alert.Tags(map[string]string{
			"service":                  "wasp",
			DefaultRequirementLabelKey: a.RequirementGroupName,
		}),
		alert.WithLokiQuery(
			a.Name,
			LokiAlertParams(a.AlertType, a.TestName, a.GenName),
		),
		alert.If(alert.Last, a.Name, a.AlertIf),
		alert.EvaluateEvery(DefaultAlertEvaluateEvery),
	)
}

// defaultLabelValuesVar creates a dashboard variable with All/Multiple options
func defaultLabelValuesVar(name, datasourceName string) dashboard.Option {
	return dashboard.VariableAsQuery(
		name,
		query.DataSource(datasourceName),
		query.Multiple(),
		query.IncludeAll(),
		query.Request(fmt.Sprintf("label_values(%s)", name)),
		query.Sort(query.NumericalAsc),
	)
}

// timeSeriesWithAlerts creates timeseries graphs per alert + definition of alert
func timeSeriesWithAlerts(datasourceName string, alertDefs []WaspAlert) []dashboard.Option {
	dashboardOpts := make([]dashboard.Option, 0)
	for _, a := range alertDefs {
		// for wasp metrics we also create additional row per alert
		tsOpts := []timeseries.Option{
			timeseries.Transparent(),
			timeseries.Span(12),
			timeseries.Height("200px"),
			timeseries.DataSource(datasourceName),
			timeseries.Legend(timeseries.Bottom),
		}
		tsOpts = append(tsOpts, defaultLastValueAlertWidget(a))

		var rowTitle string
		// for wasp metrics we also create additional row per alert
		if a.CustomAlert == nil {
			rowTitle = fmt.Sprintf("Alert: %s, Requirement: %s", a.Name, a.RequirementGroupName)
			tsOpts = append(tsOpts, timeseries.WithPrometheusTarget(LokiAlertParams(a.AlertType, a.TestName, a.GenName)))
		} else {
			rowTitle = fmt.Sprintf("External alert: %s, Requirement: %s", a.Name, a.RequirementGroupName)
		}
		// all the other custom alerts may burden the dashboard,
		dashboardOpts = append(dashboardOpts,
			dashboard.Row(
				rowTitle,
				row.Collapse(),
				row.HideTitle(),
				row.WithTimeSeries(a.Name, tsOpts...),
			))
	}
	return dashboardOpts
}

// dashboard is internal appendable representation of all Dashboard widgets
func (m *Dashboard) dashboard(datasourceName string, requirements []WaspAlert) []dashboard.Option {
	do := []dashboard.Option{
		dashboard.UID(DefaultDashboardUUID),
		dashboard.AutoRefresh("5"),
		dashboard.Time("now-30m", "now"),
		dashboard.Tags([]string{"generated"}),
		dashboard.TagsAnnotation(dashboard.TagAnnotation{
			Name:       "LoadTesting",
			Datasource: "-- Grafana --",
			IconColor:  "#5794F2",
			Tags:       []string{"load-testing"},
		}),
		defaultLabelValuesVar("go_test_name", datasourceName),
		defaultLabelValuesVar("gen_name", datasourceName),
		defaultLabelValuesVar("branch", datasourceName),
		defaultLabelValuesVar("commit", datasourceName),
		dashboard.Row(
			"Load stats",
			defaultStatWidget(
				"RPS",
				datasourceName,
				`
max_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap current_rps [$__range]) by (go_test_name, gen_name)`,
				`{{go_test_name}} {{gen_name}} RPS`,
			),
			defaultStatWidget(
				"Instances",
				datasourceName,
				`
max_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap current_instances [$__range]) by (go_test_name, gen_name)
`,
				`{{go_test_name}} {{gen_name}} Instances`,
			),
			defaultStatWidget(
				"Responses/sec",
				datasourceName,
				`
count_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} [1s])
`,
				`{{go_test_name}} {{gen_name}} Responses/sec`,
			),
			defaultStatWidget(
				"Successful requests",
				datasourceName,
				`
max_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap success [$__range]) by (go_test_name, gen_name)
`,
				`{{go_test_name}} {{gen_name}} Successful requests`,
			),
			defaultStatWidget(
				"Errored requests",
				datasourceName,
				`
max_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap failed [$__range]) by (go_test_name, gen_name)
`,
				`{{go_test_name}} {{gen_name}} Errored requests`,
			),
			defaultStatWidget(
				"Timed out requests",
				datasourceName,
				`
max_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap callTimeout [$__range]) by (go_test_name, gen_name)
`,
				`{{go_test_name}} {{gen_name}} Timed out requests`,
			),
			row.WithTimeSeries(
				"Target RPS per stages",
				timeseries.Legend(timeseries.Hide),
				timeseries.Transparent(),
				timeseries.Span(6),
				timeseries.Height("300px"),
				timeseries.DataSource(datasourceName),
				timeseries.WithPrometheusTarget(
					`
last_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap current_rps[$__interval])
`,
				),
				timeseries.WithPrometheusTarget(
					`
last_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap current_instances[$__interval])
`,
				),
			),
			row.WithTimeSeries(
				"Responses/sec",
				timeseries.Legend(timeseries.Hide),
				timeseries.Transparent(),
				timeseries.Span(6),
				timeseries.Height("300px"),
				timeseries.DataSource(datasourceName),
				timeseries.Axis(
					axis.Unit("Responses"),
					axis.Label("Responses"),
				),
				timeseries.Legend(timeseries.Bottom),
				timeseries.WithPrometheusTarget(
					`
count_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} [1s])
`, prometheus.Legend("{{go_test_name}} {{gen_name}} responses/sec"),
				),
			),
			row.WithTimeSeries(
				"Latency quantiles over groups (99, 50)",
				timeseries.Legend(timeseries.Hide),
				timeseries.Transparent(),
				timeseries.Span(6),
				timeseries.Height("300px"),
				timeseries.DataSource(datasourceName),
				timeseries.Legend(timeseries.Bottom),
				timeseries.Axis(
					axis.Unit("ms"),
					axis.Label("ms"),
				),
				timeseries.WithPrometheusTarget(
					`
quantile_over_time(0.99, {go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{go_test_name}} {{gen_name}} Q 99 - {{error}}"),
				),
				timeseries.WithPrometheusTarget(
					`
quantile_over_time(0.95, {go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{go_test_name}} {{gen_name}} Q 95 - {{error}}"),
				),
				timeseries.WithPrometheusTarget(
					`
quantile_over_time(0.50, {go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{go_test_name}} {{gen_name}} Q 50 - {{error}}"),
				),
			),
			row.WithTimeSeries(
				"Responses latency by types over time",
				timeseries.Legend(timeseries.Hide),
				timeseries.Transparent(),
				timeseries.Span(6),
				timeseries.Height("300px"),
				timeseries.DataSource(datasourceName),
				timeseries.Axis(
					axis.Unit("ms"),
					axis.Label("ms"),
				),
				timeseries.Legend(timeseries.Bottom),
				timeseries.WithPrometheusTarget(
					`
last_over_time({go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{go_test_name}} {{gen_name}} timeout: {{timeout}} errored: {{error}}"),
				),
			),
		),
		dashboard.Row(
			"Debug data",
			row.Collapse(),
			row.WithStat(
				"Latest stage stats",
				stat.Transparent(),
				stat.DataSource(datasourceName),
				stat.Text(stat.TextValueAndName),
				stat.SparkLine(),
				stat.Span(12),
				stat.Height("100px"),
				stat.ColorValue(),
				stat.WithPrometheusTarget(`
sum(bytes_over_time({go_test_name=~"${go_test_name:pipe}", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} [$__range]) * 1e-6)
`, prometheus.Legend("Overall logs size")),
				stat.WithPrometheusTarget(`
sum(bytes_rate({go_test_name=~"${go_test_name:pipe}", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} [$__interval]) * 1e-6)
`, prometheus.Legend("Logs size per second")),
			),
			row.WithLogs(
				"Stats logs",
				logs.DataSource(datasourceName),
				logs.Span(12),
				logs.Height("300px"),
				logs.Transparent(),
				logs.WithLokiTarget(`
{go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
`),
			),
			row.WithLogs(
				"Failed responses",
				logs.DataSource(datasourceName),
				logs.Span(6),
				logs.Height("300px"),
				logs.Transparent(),
				logs.WithLokiTarget(`
{go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} |~ "failed\":true"`),
			),
			row.WithLogs(
				"Timed out responses",
				logs.DataSource(datasourceName),
				logs.Span(6),
				logs.Height("300px"),
				logs.Transparent(),
				logs.WithLokiTarget(`
{go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} |~ "timeout\":true"`),
			),
		),
	}
	return append(do, timeSeriesWithAlerts(datasourceName, requirements)...)
}

// Dashboard creates dashboard instance
func (m *Dashboard) Dashboard(datasourceName string, requirements []WaspAlert) (dashboard.Builder, error) {
	return dashboard.New(
		"Wasp load generator",
		m.dashboard(datasourceName, requirements)...,
	)
}
