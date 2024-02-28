package dashboard

import (
	"github.com/K-Phoen/grabana/row"
	"github.com/K-Phoen/grabana/target/prometheus"
	"github.com/K-Phoen/grabana/timeseries"
	"github.com/K-Phoen/grabana/timeseries/axis"
)

func RPSPanel(dataSource string, labels map[string]string) row.Option {
	labelString := ""
	for key, value := range labels {
		labelString += key + "=\"" + value + "\", "
	}
	return row.WithTimeSeries(
		"Responses/sec (Generator, CallGroup)",
		timeseries.Legend(timeseries.Hide),
		timeseries.Transparent(),
		timeseries.Span(6),
		timeseries.Height("300px"),
		timeseries.DataSource(dataSource),
		timeseries.Axis(
			axis.Unit("Responses"),
			axis.Label("Responses"),
		),
		timeseries.Legend(timeseries.Bottom),
		timeseries.WithPrometheusTarget(
			`sum(count_over_time({`+labelString+`go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}", call_group=~"${call_group:pipe}"} [1s])) by (node_id, go_test_name, gen_name, call_group)`,
			prometheus.Legend("{{go_test_name}} {{gen_name}} {{call_group}} responses/sec"),
		),
		timeseries.WithPrometheusTarget(
			`sum(count_over_time({`+labelString+`go_test_name=~"${go_test_name:pipe}", test_data_type=~"responses", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"} [1s])) by (node_id, go_test_name, gen_name)`,
			prometheus.Legend("{{go_test_name}} Total responses/sec"),
		),
	)
}

func RPSVUPerScheduleSegmentsPanel(dataSource string, labels map[string]string) row.Option {
	labelString := ""
	for key, value := range labels {
		labelString += key + "=\"" + value + "\", "
	}
	return row.WithTimeSeries(
		"RPS/VUs per schedule segments",
		timeseries.Transparent(),
		timeseries.Span(6),
		timeseries.Height("300px"),
		timeseries.DataSource(dataSource),
		timeseries.WithPrometheusTarget(
			`
			max_over_time({`+labelString+`go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
			| json
			| unwrap current_rps [$__interval]) by (node_id, go_test_name, gen_name)
			`, prometheus.Legend("{{go_test_name}} {{gen_name}} RPS"),
		),
		timeseries.WithPrometheusTarget(
			`
			sum(last_over_time({`+labelString+`go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
			| json
			| unwrap current_rps [$__interval]) by (node_id, go_test_name, gen_name))
			`,
			prometheus.Legend("{{go_test_name}} Total RPS"),
		),
		timeseries.WithPrometheusTarget(
			`
			max_over_time({`+labelString+`go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
			| json
			| unwrap current_instances [$__interval]) by (node_id, go_test_name, gen_name)
			`, prometheus.Legend("{{go_test_name}} {{gen_name}} VUs"),
		),
		timeseries.WithPrometheusTarget(
			`
			sum(last_over_time({`+labelString+`go_test_name=~"${go_test_name:pipe}", test_data_type=~"stats", branch=~"${branch:pipe}", commit=~"${commit:pipe}", gen_name=~"${gen_name:pipe}"}
			| json
			| unwrap current_instances [$__interval]) by (node_id, go_test_name, gen_name))
			`,
			prometheus.Legend("{{go_test_name}} Total VUs"),
		),
	)
}
