package main

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

var NebiusObservability = dashboard.NewDashboardBuilder("Nebius Observability Platform").
	Uid("nebius-observability").
	Tags([]string{"Nebius", "Observability Platform"}).
	Refresh("1m").
	Time("now-1h", "now").
	Timezone("browser").
	Readonly().
	Tooltip(dashboard.DashboardCursorSyncCrosshair).
	WithVariable(DatasourceVar).
	Description("Unified overview of Nebius Observability usage. https://docs.nebius.com/observability").
	Link(dashboard.NewDashboardLinkBuilder("Docs").
		Type(dashboard.DashboardLinkTypeLink).
		Url("https://docs.nebius.com/observability").
		TargetBlank(true).
		Icon("doc"),
	).
	Link(dashboard.NewDashboardLinkBuilder("GitHub").
		Type(dashboard.DashboardLinkTypeLink).
		Url("https://github.com/nebius/observability").
		TargetBlank(true).
		Icon("external link"),
	).
	WithRow(
		dashboard.NewRowBuilder("Monitoring").
			GridPos(dashboard.GridPos{H: 1, W: 24, X: 0, Y: 0}).
			Id(100),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Write requests").
			Datasource(DatasourceRef).
			Description("Number of metrics ingestion requests per second").
			Unit("reqps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 0, Y: 1}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum(rate(requests_total{type="write"}[$__rate_interval])) or on() vector(0)`).
					LegendFormat("Requests").
					RefId("Requests rate").
					Range(),
			).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`requests_limits{type="monitoring.write.throughput.requests"}`).
					LegendFormat("Requests limit").
					RefId("Requests limit").
					Range(),
			).
			OverrideByQuery("Requests limit", []dashboard.DynamicConfigValue{
				fixedColor("dark-red"),
				{Id: "custom.fillOpacity", Value: 0},
				{Id: "custom.hideFrom", Value: map[string]bool{"legend": true, "tooltip": false, "viz": false}},
				{Id: "custom.drawStyle", Value: "line"},
				{Id: "custom.showPoints", Value: "never"},
				{Id: "custom.lineWidth", Value: 1},
			}).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithPanel(
		applyHttpStatusOverrides(
			timeseries.NewPanelBuilder().
				Title("Write errors").
				Datasource(DatasourceRef).
				Description("Number of failed metrics ingestion requests per second, by status code").
				Unit("reqps").
				GridPos(dashboard.GridPos{H: 8, W: 12, X: 12, Y: 1}).
				WithTarget(
					prometheus.NewDataqueryBuilder().
						Expr(`sum by (status_code) (rate(requests_total{status_code!~"2.*", type="write"}[$__rate_interval])) or on() vector(0)`).
						LegendFormat("{{status_code}}").
						RefId("Write errors by status").
						Range(),
				),
		),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Read requests").
			Datasource(DatasourceRef).
			Description("Number of metrics read requests per second").
			Unit("reqps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 0, Y: 9}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum(rate(requests_total{type="read"}[$__rate_interval])) or on() vector(0)`).
					LegendFormat("Requests").
					RefId("Requests").
					Range(),
			).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`requests_limits{type="monitoring.read.throughput.requests"}`).
					LegendFormat("Limit").
					RefId("Limit").
					Range(),
			).
			OverrideByQuery("Limit", []dashboard.DynamicConfigValue{
				fixedColor("dark-red"),
				{Id: "custom.drawStyle", Value: "line"},
				{Id: "custom.showPoints", Value: "never"},
				{Id: "custom.lineWidth", Value: 1},
			}).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithPanel(
		applyHttpStatusOverrides(
			timeseries.NewPanelBuilder().
				Title("Read errors").
				Datasource(DatasourceRef).
				Description("Number of failed metrics read requests per second, by status code").
				Unit("reqps").
				GridPos(dashboard.GridPos{H: 8, W: 12, X: 12, Y: 9}).
				WithTarget(
					prometheus.NewDataqueryBuilder().
						Expr(`sum by (status_code) (rate(requests_total{status_code!~"2.*", type="read"}[$__rate_interval])) or on() vector(0)`).
						LegendFormat("{{status_code}}").
						RefId("Read errors by status").
						Range(),
				),
		),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Samples write rate").
			Datasource(DatasourceRef).
			Description("Number of samples ingested per second, by type").
			Unit("rowsps").
			GridPos(dashboard.GridPos{H: 8, W: 24, X: 0, Y: 17}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum by(type) (rate(samples_total{}[10m])) OR on() vector(0)`).
					LegendFormat("{{type}}").
					RefId("Samples rate by type").
					Range(),
			).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithRow(
		dashboard.NewRowBuilder("Logging").
			GridPos(dashboard.GridPos{H: 1, W: 24, X: 0, Y: 25}).
			Id(200),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Write requests").
			Datasource(DatasourceRef).
			Description("Number of successful log ingestion requests per second").
			Unit("reqps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 0, Y: 26}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum(rate(logging_ingest_requests_total{status="ok"}[$__rate_interval])) OR on() vector(0)`).
					LegendFormat("Requests").
					RefId("A").
					Range(),
			).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Write errors").
			Datasource(DatasourceRef).
			Description("Number of failed log ingestion requests per second, by status code").
			Unit("reqps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 12, Y: 26}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum by(status) (rate(logging_ingest_requests_total{status!="ok"}[$__rate_interval])) or on() vector(0)`).
					LegendFormat("{{status}}").
					RefId("A").
					Range(),
			).
			OverrideByName("err_auth", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Auth error"}, fixedColor("#FFF176")}).
			OverrideByName("err_process", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Processing error"}, fixedColor("#FFB3B8")}).
			OverrideByName("err_validate", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Validation error"}, fixedColor("#FF9E80")}).
			OverrideByName("quota_exceeded", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Quota exceeded"}, fixedColor("#FFB347")}).
			OverrideByName("workspace_inactive", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Inactive workspace"}, fixedColor("#FFD966")}).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Read requests").
			Datasource(DatasourceRef).
			Description("Number of successful log read/query requests per second").
			Unit("reqps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 0, Y: 34}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum(rate(logging_read_requests_total{status="ok"}[$__rate_interval])) OR on() vector(0)`).
					LegendFormat("Requests").
					RefId("A").
					Range(),
			).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Read errors").
			Datasource(DatasourceRef).
			Description("Number of failed log read/query requests per second, by status code").
			Unit("reqps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 12, Y: 34}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum by(status) (rate(logging_read_requests_total{status!="ok"}[$__rate_interval])) or on() vector(0)`).
					LegendFormat("{{status}}").
					RefId("A").
					Range(),
			).
			OverrideByName("err_auth", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Auth error"}, fixedColor("#FFF176")}).
			OverrideByName("err_process", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Processing error"}, fixedColor("#FFB3B8")}).
			OverrideByName("err_validate", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Validation error"}, fixedColor("#FF9E80")}).
			OverrideByName("quota_exceeded", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Quota exceeded"}, fixedColor("#FFB347")}).
			OverrideByName("workspace_inactive", []dashboard.DynamicConfigValue{{Id: "displayName", Value: "Inactive workspace"}, fixedColor("#FFD966")}).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{
					{Id: "custom.drawStyle", Value: "line"},
					{Id: "custom.showPoints", Value: "never"},
					{Id: "custom.lineWidth", Value: 1},
				},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Written lines rate").
			Datasource(DatasourceRef).
			Description("Number of log lines ingested per second").
			Unit("short").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 0, Y: 42}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum(rate(logging_ingest_logs_total{}[$__rate_interval])) OR on() vector(0)`).
					LegendFormat("Lines").
					RefId("A").
					Range(),
			).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1}},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Write bytes").
			Datasource(DatasourceRef).
			Description("Volume of log data ingested per second in bytes").
			Unit("binBps").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 12, Y: 42}).
			WithTarget(
				prometheus.NewDataqueryBuilder().
					Expr(`sum(rate(logging_ingest_logs_bytes_total{}[$__rate_interval])) OR on() vector(0)`).
					LegendFormat("Data").
					RefId("A").
					Range(),
			).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1}},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Write duration (p50/p75/p90/p95/p99)").
			Datasource(DatasourceRef).
			Description("Request processing time quantiles for log ingestion operations").
			Unit("s").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 0, Y: 50}).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.5, sum by(le)(rate(logging_ingest_duration_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p50").RefId("A").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.75, sum by(le)(rate(logging_ingest_duration_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p75").RefId("B").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.90, sum by(le)(rate(logging_ingest_duration_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p90").RefId("C").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.95, sum by(le)(rate(logging_ingest_duration_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p95").RefId("D").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.99, sum by(le)(rate(logging_ingest_duration_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p99").RefId("E").Range()).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1}},
			),
	).
	WithPanel(
		timeseries.NewPanelBuilder().
			Title("Logs save lag (p50/p75/p90/p95/p99)").
			Datasource(DatasourceRef).
			Description("Time delay between receiving a log and saving it to storage").
			Unit("s").
			GridPos(dashboard.GridPos{H: 8, W: 12, X: 12, Y: 50}).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.5, sum by(le)(rate(logging_storage_save_lag_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p50").RefId("A").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.75, sum by(le)(rate(logging_storage_save_lag_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p75").RefId("B").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.90, sum by(le)(rate(logging_storage_save_lag_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p90").RefId("C").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.95, sum by(le)(rate(logging_storage_save_lag_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p95").RefId("D").Range()).
			WithTarget(prometheus.NewDataqueryBuilder().Expr(`histogram_quantile(0.99, sum by(le)(rate(logging_storage_save_lag_seconds_bucket{}[$__rate_interval])))`).LegendFormat("p99").RefId("E").Range()).
			WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
				[]dashboard.DynamicConfigValue{{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1}},
			),
	)

func fixedColor(col string) dashboard.DynamicConfigValue {
	c, _ := dashboard.NewFieldColorBuilder().
		Mode(dashboard.FieldColorModeId("fixed")).
		FixedColor(col).
		Build()
	return dashboard.DynamicConfigValue{Id: "color", Value: c}
}

func applyHttpStatusOverrides(p *timeseries.PanelBuilder) *timeseries.PanelBuilder {
	return p.
		OverrideByName("400", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "400: Bad Request"}, fixedColor("#FFF176"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("401", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "401: Unauthorized"}, fixedColor("#FFB3B8"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("403", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "403: Forbidden"}, fixedColor("#FF9E80"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("404", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "404: Not Found"}, fixedColor("#FFB347"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("408", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "408: Request Timeout"}, fixedColor("#FFD966"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("409", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "409: Conflict"}, fixedColor("#FFE066"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("412", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "412: Invalid token"}, fixedColor("#FFE084"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("422", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "422: Unprocessable Entity"}, fixedColor("#FFC1E3"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("429", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "429: Too Many Requests"}, fixedColor("#FFEE58"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("502", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "502: Bad Gateway"}, fixedColor("#FF8A80"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("503", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "503: Service Unavailable"}, fixedColor("#FF5252"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		OverrideByName("504", []dashboard.DynamicConfigValue{
			{Id: "displayName", Value: "504: Gateway Timeout"}, fixedColor("#E57373"),
			{Id: "custom.drawStyle", Value: "line"}, {Id: "custom.showPoints", Value: "never"}, {Id: "custom.lineWidth", Value: 1},
		}).
		WithOverride(dashboard.MatcherConfig{Id: "byRegexp", Options: ".*"},
			[]dashboard.DynamicConfigValue{
				{Id: "custom.drawStyle", Value: "line"},
				{Id: "custom.showPoints", Value: "never"},
				{Id: "custom.lineWidth", Value: 1},
			},
		)
}
