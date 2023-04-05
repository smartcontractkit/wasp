package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/K-Phoen/grabana"
	"github.com/K-Phoen/grabana/dashboard"
	"github.com/K-Phoen/grabana/logs"
	"github.com/K-Phoen/grabana/row"
	"github.com/K-Phoen/grabana/stat"
	"github.com/K-Phoen/grabana/target/prometheus"
	"github.com/K-Phoen/grabana/timeseries"
	"github.com/K-Phoen/grabana/timeseries/axis"
	"github.com/K-Phoen/grabana/variable/query"
)

// WaspDashboard is a Wasp dashboard
type WaspDashboard struct{}

// NewWaspDashboard creates new dashboard instance
func NewWaspDashboard() *WaspDashboard {
	return &WaspDashboard{}
}

// Deploy deploys this dashboard to some Grafana folder
func (m *WaspDashboard) Deploy(dsName, folder, url, token string) (*grabana.Dashboard, error) {
	ctx := context.Background()
	d, err := m.Dashboard(dsName)
	if err != nil {
		return nil, fmt.Errorf("failed to build dashboard: %s", err)
	}
	client := grabana.NewClient(&http.Client{}, url, grabana.WithAPIToken(token))
	fo, err := client.FindOrCreateFolder(ctx, folder)
	if err != nil {
		fmt.Printf("Could not find or create folder: %s\n", err)
		os.Exit(1)
	}
	return client.UpsertDashboard(ctx, fo, d)
}

// Dashboard creates dashboard instance
func (m *WaspDashboard) Dashboard(datasourceName string) (dashboard.Builder, error) {
	return dashboard.New(
		"Wasp load generator",
		dashboard.UID("wasp"),
		dashboard.AutoRefresh("5"),
		dashboard.Time("now-30m", "now"),
		dashboard.Tags([]string{"generated"}),
		dashboard.TagsAnnotation(dashboard.TagAnnotation{
			Name:       "LoadTesting",
			Datasource: "-- Grafana --",
			IconColor:  "#5794F2",
			Tags:       []string{"load-testing"},
		}),
		dashboard.VariableAsQuery(
			"test_group",
			query.DataSource(datasourceName),
			query.Request("label_values(test_group)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"app",
			query.DataSource(datasourceName),
			query.Request("label_values(app)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"cluster",
			query.DataSource(datasourceName),
			query.Request("label_values(cluster)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"namespace",
			query.DataSource(datasourceName),
			query.Request("label_values(namespace)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"go_test_name",
			query.DataSource(datasourceName),
			query.Request("label_values(go_test_name)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"branch",
			query.DataSource(datasourceName),
			query.Request("label_values(branch)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"commit",
			query.DataSource(datasourceName),
			query.Request("label_values(commit)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.VariableAsQuery(
			"test_id",
			query.DataSource(datasourceName),
			query.Request("label_values(test_id)"),
			query.Sort(query.NumericalAsc),
		),
		dashboard.Row(
			"Load stats",
			row.WithStat(
				"Latest stage stats",
				stat.Transparent(),
				stat.DataSource(datasourceName),
				stat.Text(stat.TextValueAndName),
				stat.Span(12),
				stat.Height("100px"),
				stat.ColorValue(),
				stat.WithPrometheusTarget(`
max_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap current_rps [$__range]) by (test_id)
`, prometheus.Legend("{{test_id}} Target RPS")),
				stat.WithPrometheusTarget(`
max_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap current_instances [$__range]) by (test_id)
`, prometheus.Legend("{{test_id}} Instances")),
				stat.WithPrometheusTarget(`
count_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"} [1s])
`, prometheus.Legend("{{test_id}} Responses/sec")),
				stat.WithPrometheusTarget(`
max_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap success [$__range]) by (test_id)
`, prometheus.Legend("{{test_id}} Successful requests")),
				stat.WithPrometheusTarget(`
max_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="$test_group", test_id=~"${test_id:pipe}"}
| json
| unwrap failed [$__range]) by (test_id)
`, prometheus.Legend("{{test_id}} Errored requests")),
				stat.WithPrometheusTarget(`
max_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="$test_group", test_id=~"${test_id:pipe}"}
| json
| unwrap callTimeout [$__range]) by (test_id)
`, prometheus.Legend("{{test_id}} Timed out requests")),
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
last_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap current_rps[$__interval])
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
count_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"} [1s])
`, prometheus.Legend("{{test_id}} responses/sec"),
				),
			),
			row.WithTimeSeries(
				"Latency quantiles over groups (99,  50)",
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
quantile_over_time(0.99, {cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{test_id}} Q 99 - {{error}}"),
				),
				timeseries.WithPrometheusTarget(
					`
quantile_over_time(0.95, {cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{test_id}} Q 95 - {{error}}"),
				),
				timeseries.WithPrometheusTarget(
					`
quantile_over_time(0.50, {cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{test_id}} Q 50 - {{error}}"),
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
last_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="${test_group}", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
| json
| unwrap duration [$__interval]) / 1e6
`, prometheus.Legend("{{test_id}} timeout: {{timeout}} errored: {{error}}"),
				),
			),
		),
		dashboard.Row(
			"Debug data",
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
sum(bytes_over_time({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"} [$__range]) * 1e-6)
`, prometheus.Legend("Overall logs size")),
				stat.WithPrometheusTarget(`
sum(bytes_rate({cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"} [$__interval]) * 1e-6)
`, prometheus.Legend("Logs size per second")),
			),
			row.WithLogs(
				"Stats logs",
				logs.DataSource(datasourceName),
				logs.Span(12),
				logs.Height("300px"),
				logs.Transparent(),
				logs.WithLokiTarget(`
{cluster="${cluster}", namespace="${namespace}", app="${app}", go_test_name="${go_test_name:pipe}", test_data_type="stats", test_group="${test_group}", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"}
`),
			),
			row.WithLogs(
				"Failed responses",
				logs.DataSource(datasourceName),
				logs.Span(6),
				logs.Height("300px"),
				logs.Transparent(),
				logs.WithLokiTarget(`
{cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"} |= "failed\":true"`),
			),
			row.WithLogs(
				"Timed out responses",
				logs.DataSource(datasourceName),
				logs.Span(6),
				logs.Height("300px"),
				logs.Transparent(),
				logs.WithLokiTarget(`
{cluster="${cluster}", app="${app}", namespace="${namespace}", go_test_name="${go_test_name:pipe}", test_data_type="responses", test_group="$test_group", test_id=~"${test_id:pipe}", branch="${branch:pipe}", commit="${commit:pipe}"} |= "timeout\":true"`),
			),
		),
	)
}

func main() {
	if _, err := NewWaspDashboard().Deploy(
		os.Getenv("DATA_SOURCE_NAME"),
		os.Getenv("DASHBOARD_FOLDER"),
		os.Getenv("GRAFANA_URL"),
		os.Getenv("GRAFANA_TOKEN"),
	); err != nil {
		panic(err)
	}
}
