package main

import (
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
	"github.com/grafana/grafana-foundation-sdk/go/units"
)

var NebiusObjectStorage = dashboard.NewDashboardBuilder("Nebius Object Storage").
	Uid("nebius-object-storage").
	Description("Nebius Object Storage Overview.").
	Tags([]string{"Nebius", "Object Storage"}).
	// Links (match new JSON)
	Link(dashboard.NewDashboardLinkBuilder("Docs").
		Type(dashboard.DashboardLinkTypeLink).
		Url("https://docs.nebius.com/object-storage").
		TargetBlank(true).
		Icon("doc"),
	).
	Link(dashboard.NewDashboardLinkBuilder("GitHub").
		Type(dashboard.DashboardLinkTypeLink).
		Url("https://github.com/nebius/observability").
		TargetBlank(true).
		Icon("external link"),
	).
	WithVariable(DatasourceVar).
	WithVariable(
		dashboard.NewQueryVariableBuilder("bucket").
			Datasource(DatasourceRef).
			Query(dashboard.StringOrMap{
				String: New("label_values(buckets_stat_quantity, bucket)"),
			}).
			Multi(true).
			AllowCustomValue(false).
			IncludeAll(true).
			AllValue(".*"),
	).

	// ─────────────────────────────────────────────────────────────────────────────
	// Storage space row
	// ─────────────────────────────────────────────────────────────────────────────
	WithRow(
		dashboard.NewRowBuilder("Storage space"),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Traffic").
		Description("Data transfer speed to and from storage.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (rate(http_bytes_sent{bucket=~"$bucket"}[$__rate_interval])) OR on() vector(0)`).
			LegendFormat("Download {{bucket}}").
			RefId("A"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (rate(http_bytes_received{bucket=~"$bucket"}[$__rate_interval])) OR on() vector(0)`).
			LegendFormat("Upload {{bucket}}").
			RefId("B"),
		).
		Unit(units.BytesPerSecondIEC).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Tooltip(common.NewVizTooltipOptionsBuilder().
			Mode(common.TooltipDisplayModeMulti).
			Sort(common.SortOrderNone),
		).
		Legend(common.NewVizLegendOptionsBuilder().
			ShowLegend(true).
			Placement(common.LegendPlacementBottom),
		).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		OverrideByName("Download ", []dashboard.DynamicConfigValue{
			{
				Id: "custom.hideFrom",
				Value: map[string]interface{}{
					"legend":  true,
					"tooltip": true,
					"viz":     false,
				},
			},
		}).
		OverrideByName("Upload ", []dashboard.DynamicConfigValue{
			{
				Id: "custom.hideFrom",
				Value: map[string]interface{}{
					"legend":  true,
					"tooltip": true,
					"viz":     true,
				},
			},
		}).
		Height(9).
		Span(8),
	).

	// Total bucket size
	WithPanel(timeseries.NewPanelBuilder().
		Title("Total bucket size").
		Description("Storage space used by all objects in a bucket.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket"}[1m])))`).
			LegendFormat("{{bucket}}").
			RefId("A"),
		).
		Unit(units.BytesIEC).
		FillOpacity(70).
		ShowPoints(common.VisibilityModeNever).
		Tooltip(common.NewVizTooltipOptionsBuilder().
			Mode(common.TooltipDisplayModeSingle).
			Sort(common.SortOrderNone),
		).
		Legend(common.NewVizLegendOptionsBuilder().
			ShowLegend(true).
			Placement(common.LegendPlacementBottom),
		).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(9).
		Span(8),
	).

	// Space by storage class
	WithPanel(timeseries.NewPanelBuilder().
		Title("Space by storage class").
		Description("Amount of storage used by objects in different storage classes.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket",storage_class="STANDARD"}[1m])))`).
			LegendFormat("{{bucket}} Standard").
			RefId("A"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket",storage_class="ENHANCED_THROUGHPUT"}[1m])))`).
			LegendFormat("{{bucket}} Enhanced Throughput").
			RefId("B"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket",storage_class="INTELLIGENT"}[1m])))`).
			LegendFormat("{{bucket}} Intelligent ({{intelligent_tier}})").
			RefId("C"),
		).
		Unit(units.BytesIEC).
		FillOpacity(70).
		ShowPoints(common.VisibilityModeNever).
		Tooltip(common.NewVizTooltipOptionsBuilder().
			Mode(common.TooltipDisplayModeSingle).
			Sort(common.SortOrderNone),
		).
		Legend(common.NewVizLegendOptionsBuilder().
			ShowLegend(true).
			Placement(common.LegendPlacementBottom),
		).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(9).
		Span(8),
	).

	// ─────────────────────────────────────────────────────────────────────────────
	// Requests row (repeated per bucket)
	// ─────────────────────────────────────────────────────────────────────────────
	WithRow(
		dashboard.NewRowBuilder("Requests for $bucket").
			Repeat("bucket"),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Read requests").
		Description("Number of requests made to retrieve object content from a bucket. ").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(handler) (rate(request_rate{bucket="$bucket", operation_type="read"}[$__rate_interval]))`).
			LegendFormat("{{handler}}").
			RefId("A"),
		).
		Unit(units.RequestsPerSecond).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(8),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Modify requests").
		Description("Number of requests made to upload objects or modify object content. ").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(handler) (rate(request_rate{bucket="$bucket", operation_type="mutate"}[$__rate_interval]))`).
			LegendFormat("{{handler}}").
			RefId("A"),
		).
		Unit(units.RequestsPerSecond).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(8),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("API errors").
		Description("Number of errors when accessing S3 API. Number of errors per 5 minutes.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(handler, http_code, api_error_code) (increase(http_errors_total{bucket="$bucket"}[5m]))`).
			LegendFormat("{{handler}}:{{http_code}}:{{api_error_code}}").
			RefId("A"),
		).
		Unit(units.Short).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(8),
	).

	// ─────────────────────────────────────────────────────────────────────────────
	// Objects statistics row
	// ─────────────────────────────────────────────────────────────────────────────
	WithRow(
		dashboard.NewRowBuilder("Objects statistics"),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Objects count").
		Description("Number of objects. Single, multipart objects, and incomplete multipart uploads are counted separately.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", counter="simple_objects"}[1m])))`).
			LegendFormat("{{bucket}} Simple objects").
			RefId("A"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", counter="multipart_objects"}[1m])))`).
			LegendFormat("{{bucket}} Multipart objects").
			RefId("B"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", counter="inflight_parts"}[1m])))`).
			LegendFormat("{{bucket}} Multipart uploads").
			RefId("C"),
		).
		Unit(units.Short).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(8),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Space by object type").
		Description("Amount of storage used by objects of different types.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket", counter="simple_objects"}[1m])))`).
			LegendFormat("{{bucket}} Simple objects").
			RefId("A"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket", counter="multipart_objects"}[1m])))`).
			LegendFormat("{{bucket}} Multipart objects").
			RefId("B"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket", counter="inflight_parts"}[1m])))`).
			LegendFormat("{{bucket}} Multipart uploads").
			RefId("C"),
		).
		Unit(units.BytesIEC).
		FillOpacity(70).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(8),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Objects count by storage class").
		Description("Number of objects stored in different storage classes.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", storage_class="STANDARD"}[1m])))`).
			LegendFormat("{{bucket}} Standard").
			RefId("A"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", storage_class="ENHANCED_THROUGHPUT"}[1m])))`).
			LegendFormat("{{bucket}} Enhanced Throughput").
			RefId("B"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", storage_class="INTELLIGENT"}[1m])))`).
			LegendFormat("{{bucket}} Intelligent ({{intelligent_tier}})").
			RefId("C"),
		).
		Unit(units.Short).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(8),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Space by version").
		Description("Amount of storage used by current and noncurrent object versions.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket, counters_type) (max without(iam_container_id, node) (last_over_time(buckets_stat_size{bucket=~"$bucket", counter=~"simple_objects|multipart_objects"}[1m])))`).
			LegendFormat("{{bucket}} {{counters_type}}").
			RefId("A"),
		).
		Unit(units.BytesIEC).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(12),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Objects count by version").
		Description("Number of current and noncurrent object versions.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by(bucket, counters_type) (max without(iam_container_id, node) (last_over_time(buckets_stat_quantity{bucket=~"$bucket", counter=~"simple_objects|multipart_objects"}[1m])))`).
			LegendFormat("{{bucket}} {{counters_type}}").
			RefId("A"),
		).
		Unit(units.Short).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(12),
	).

	// ─────────────────────────────────────────────────────────────────────────────
	// Lifecycle row
	// ─────────────────────────────────────────────────────────────────────────────
	WithRow(
		dashboard.NewRowBuilder("Lifecycle"),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Traffic by Transition Lifecycle").
		Description("Traffic consumed by Transition Lifecycle of a bucket.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum by (storage_class) (rate(lifecycle_transition_network_bytes{bucket=~"$bucket", direction="traffic-in"}[1m])) > 0`).
			LegendFormat("{{bucket}} upload to {{storage_class}}").
			RefId("A"),
		).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`-sum by (storage_class) (rate(lifecycle_transition_network_bytes{bucket=~"$bucket", direction="traffic-out"}[1m])) < 0`).
			LegendFormat("{{bucket}} download from {{storage_class}}").
			RefId("B"),
		).
		Unit(units.BytesPerSecondIEC).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(12),
	).
	WithPanel(timeseries.NewPanelBuilder().
		Title("Transition Lifecycle failures").
		Description("Failures of Transition Lifecycle.").
		Datasource(DatasourceRef).
		WithTarget(prometheus.NewDataqueryBuilder().
			Expr(`sum(increase_pure(lifecycle_transition_quota_limit_exceeded_total{bucket=~"$bucket"}[1m]))`).
			LegendFormat("{{bucket}} upload to {{storage_class}}").
			RefId("A"),
		).
		Unit(units.Short).
		FillOpacity(5).
		ShowPoints(common.VisibilityModeNever).
		Thresholds(dashboard.NewThresholdsConfigBuilder()).
		Height(8).
		Span(12),
	).
	Time("now-24h", "now").
	Refresh("1m").
	Readonly()
