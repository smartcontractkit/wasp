package main

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
	"github.com/google/uuid"
)

const (
	DefaultStatTextSize       = 12
	DefaultStatValueSize      = 20
	DefaultAlertEvaluateEvery = "10s"
	DefaultAlertFor           = "10s"
	DefaultDashboardUUID      = "wasp"
)

var (
	DefaultAlertTags = map[string]string{
		"service": "wasp",
	}
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
func defaultLastValueAlertWidget(name, lokiTarget string, requirenemtName string, alertIf alert.ConditionEvaluator) timeseries.Option {
	ref := uuid.New().String()
	tags := DefaultAlertTags
	tags["requirement_name"] = requirenemtName
	return timeseries.Alert(
		name,
		alert.For(DefaultAlertFor),
		alert.OnExecutionError(alert.ErrorKO),
		alert.Description(name),
		alert.Tags(tags),
		alert.WithLokiQuery(
			ref,
			lokiTarget,
		),
		alert.If(alert.Last, ref, alertIf),
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

// Dashboard creates dashboard instance
func (m *WaspDashboard) Dashboard(datasourceName string) (dashboard.Builder, error) {
	return dashboard.New(
		"Wasp load generator",
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
				defaultLastValueAlertWidget(
					"Latency is out of SLO",
					`
avg(quantile_over_time(0.99, {go_test_name=~"TestProfile", test_data_type=~"responses"}
| json
| unwrap duration [10s]) / 1e6)
`,
					`baseline`,
					alert.IsAbove(50),
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
