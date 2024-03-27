package grafanasdk

import (
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

func RPSNowPanel(queryPrefix string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	panelName := "RPS (Now)"
	query := `sum(last_over_time({` + queryPrefix + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"}
	| json
	| unwrap current_rps [1s]) by (node_id, go_test_name, gen_name)) by (__stream_shard__)`
	legend := "{{go_test_name}} {{gen_name}} RPS"

	return defaultStatPanel(panelName, query, legend, panelID, datasource)
}

func VUsNowPanel(queryPrefix string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	panelName := "VUs (Now)"
	query := `sum(max_over_time({` + queryPrefix + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"}
	| json
	| unwrap current_instances [$__range]) by (node_id, go_test_name, gen_name)) by (__stream_shard__)`
	legend := "{{go_test_name}} {{gen_name}} VUs"

	return defaultStatPanel(panelName, query, legend, panelID, datasource)
}

func ResponsesPerSecNowPanel(queryPrefix string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	panelName := "Responses/sec (Now)"
	query := `sum(count_over_time({` + queryPrefix + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}"} [1s])) by (node_id, go_test_name, gen_name)`
	legend := "{{go_test_name}} {{gen_name}} Responses/sec"

	return defaultStatPanel(panelName, query, legend, panelID, datasource)
}

func TotalSuccessfulRequestsPanel(queryPrefix string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	panelName := "Successful requests (Total)"
	query := `
	sum(max_over_time({` + queryPrefix + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"}
	| json
	| unwrap success [$__range]) by (node_id, go_test_name, gen_name)) by (__stream_shard__)
	`
	legend := "{{go_test_name}} {{gen_name}} Successful requests"

	return defaultStatPanel(panelName, query, legend, panelID, datasource)
}

func TotalErroredRequestsPanel(queryPrefix string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	panelName := "Errored requests (Total)"
	query := `
	sum(max_over_time({` + queryPrefix + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"}
	| json
	| unwrap failed [$__range]) by (node_id, go_test_name, gen_name)) by (__stream_shard__)
	`
	legend := "{{go_test_name}} {{gen_name}} Errored requests"

	return defaultStatPanel(panelName, query, legend, panelID, datasource)
}

func TimedOutRequestsPanel(queryPrefix string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	panelName := "Timed out requests (Total)"
	query := `
	sum(max_over_time({` + queryPrefix + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"}
	| json
	| unwrap callTimeout [$__range]) by (node_id, go_test_name, gen_name)) by (__stream_shard__)
	`
	legend := "{{go_test_name}} {{gen_name}} Timed out requests"

	return defaultStatPanel(panelName, query, legend, panelID, datasource)
}

func defaultStatPanel(panelName, query, legendFormat string, panelID uint32, datasource dashboard.DataSourceRef) *stat.PanelBuilder {
	return stat.NewPanelBuilder().Title(panelName).
		Id(panelID).
		Datasource(datasource).
		Orientation(common.VizOrientationVertical).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(12).ValueSize(20)).
		TextMode(common.BigValueTextModeValueAndName).
		GraphMode(common.BigValueGraphModeNone).
		ReduceOptions(common.NewReduceDataOptionsBuilder().Calcs([]string{"last"}).Values(true)).
		// Legend(common.NewVizLegendOptionsBuilder().ShowLegend(true).Placement(common.LegendPlacementBottom).DisplayMode(common.LegendDisplayModeList).Calcs([]string{})).
		Span(4).
		Transparent(true).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(query).
				LegendFormat(legendFormat),
		)
}

func RPSPanel(queryString string, panelID uint32, promDatasource dashboard.DataSourceRef) *timeseries.PanelBuilder {
	return timeseries.NewPanelBuilder().Title("Responses/sec (Generator, CallGroup)").
		Id(panelID).
		Datasource(promDatasource).
		Height(6).
		Span(12).
		Transparent(true).
		AxisLabel("Responses").
		Legend(common.NewVizLegendOptionsBuilder().ShowLegend(false).Placement(common.LegendPlacementBottom).DisplayMode(common.LegendDisplayModeList).Calcs([]string{})).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`sum(count_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}", call_group=~"${call_group:pipe}"} [1s])) by (node_id, go_test_name, gen_name, call_group)`).
				LegendFormat("{{go_test_name}} {{gen_name}} {{call_group}} responses/sec"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`sum(count_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}"} [1s])) by (node_id, go_test_name, gen_name)`).
				LegendFormat("{{go_test_name}} Total responses/sec"),
		)
}

func RPSVUPerScheduleSegmentsPanel(queryString string, panelID uint32, datasource dashboard.DataSourceRef) *timeseries.PanelBuilder {
	return timeseries.NewPanelBuilder().Title("RPS/VUs per schedule segments").
		Id(panelID).
		Datasource(datasource).
		Height(6).
		Span(12).
		Transparent(true).
		Legend(common.NewVizLegendOptionsBuilder().ShowLegend(true).Placement(common.LegendPlacementBottom).DisplayMode(common.LegendDisplayModeList).Calcs([]string{})).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`max_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"} | json | unwrap current_rps [$__interval]) by (node_id, go_test_name, gen_name)`).
				LegendFormat("{{go_test_name}} {{gen_name}} RPS"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`sum(last_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"} | json | unwrap current_rps [$__interval]) by (node_id, go_test_name, gen_name))`).
				LegendFormat("{{go_test_name}} Total RPS"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`max_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"} | json | unwrap current_instances [$__interval]) by (node_id, go_test_name, gen_name)`).
				LegendFormat("{{go_test_name}} {{gen_name}} VUs"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`sum(last_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", gen_name=~"${gen_name:pipe}"} | json | unwrap current_instances [$__interval]) by (node_id, go_test_name, gen_name))`).
				LegendFormat("{{go_test_name}} Total VUs"),
		)
}

func LatencyQuantilesPanel(queryString string, panelID uint32, promDatasource dashboard.DataSourceRef) *timeseries.PanelBuilder {
	return timeseries.NewPanelBuilder().Title("Latency quantiles over groups (99, 95, 50)").
		Id(panelID).
		Datasource(promDatasource).
		Height(6).
		Span(12).
		Transparent(true).
		AxisLabel("ms").
		Legend(common.NewVizLegendOptionsBuilder().ShowLegend(false).Placement(common.LegendPlacementBottom).DisplayMode(common.LegendDisplayModeList).Calcs([]string{})).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`quantile_over_time(0.99, {` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}"} | json | unwrap duration [$__interval]) by (go_test_name, gen_name) / 1e6`).
				LegendFormat("{{go_test_name}} {{gen_name}} Q 99 - {{error}}"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`quantile_over_time(0.95, {` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}"} | json | unwrap duration [$__interval]) by (go_test_name, gen_name) / 1e6`).
				LegendFormat("{{go_test_name}} {{gen_name}} Q 95 - {{error}}"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`quantile_over_time(0.50, {` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}"} | json | unwrap duration [$__interval]) by (go_test_name, gen_name) / 1e6`).
				LegendFormat("{{go_test_name}} {{gen_name}} Q 50 - {{error}}"),
		)
}

func ResponseLatenciesPanel(queryString string, panelID uint32, promDatasource dashboard.DataSourceRef) *timeseries.PanelBuilder {
	return timeseries.NewPanelBuilder().Title("Responses latencies by types over time (Generator, CallGroup)").
		Id(panelID).
		Datasource(promDatasource).
		Height(6).
		Span(12).
		Transparent(true).
		AxisLabel("ms").
		Legend(common.NewVizLegendOptionsBuilder().ShowLegend(false).Placement(common.LegendPlacementBottom).DisplayMode(common.LegendDisplayModeList).Calcs([]string{})).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`last_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}", call_group=~"${call_group}"} | json | unwrap duration [$__interval]) / 1e6`).
				LegendFormat("{{go_test_name}} {{gen_name}} {{call_group}} T: {{timeout}} E: {{error}}"),
		).
		WithTarget(
			prometheus.NewDataqueryBuilder().
				Expr(`last_over_time({` + queryString + `go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", gen_name=~"${gen_name:pipe}"} | json | unwrap duration [$__interval]) / 1e6`).
				LegendFormat("{{go_test_name}} {{gen_name}} all groups T: {{timeout}} E: {{error}}"),
		)
}
