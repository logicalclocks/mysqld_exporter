// Copyright 2019, 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Scrape `ndbinfo.disk_write_speed_aggregate`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoDiskWriteSpeedAggregateQuery = `
	SELECT node_id, thr_no, backup_lcp_speed_last_10sec, redo_speed_last_10sec, 
	slowdowns_due_to_io_lag, slowdowns_due_to_high_cpu
	FROM ndbinfo.disk_write_speed_aggregate;
	`

var (
	ndbinfoDiskWriteSpeedAggregateLcpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "disk_write_speed_lcp"),
		"Number of bytes written to disk by backup and LCP processes per second, averaged over the last 10 seconds for each node and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateRedoDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "disk_write_speed_redo"),
		"Number of bytes written to REDO log processes per second, averaged over the last 10 seconds for each node and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateIOSlowdownDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "disk_write_speed_io_slowdown"),
		"Number of seconds since last node start that disk writes were slowed due to REDO log I/O lag for each node and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskWriteSpeedAggregateCPUSlowdownDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "disk_write_speed_cpu_slowdown"),
		"Number of seconds since last node start that disk writes were slowed due to high CPU usage for each node and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
)

// ScrapeNdbinfoDiskWriteSpeedAggregate collects for `ndbinfo.memoryusage`
type ScrapeNdbinfoDiskWriteSpeedAggregate struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoDiskWriteSpeedAggregate) Name() string {
	return "ndbinfo.disk_write_speed_aggregate"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoDiskWriteSpeedAggregate) Help() string {
	return "Collect metrics from ndbinfo.disk_write_speed_aggregate"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoDiskWriteSpeedAggregate) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoDiskWriteSpeedAggregate) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoDiskWriteSpeedAggregateRows, err := db.QueryContext(ctx, ndbinfoDiskWriteSpeedAggregateQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDiskWriteSpeedAggregateRows.Close()

	var (
		nodeID, threadNO, lcpWrite, redoWrite, slowdownIO, slowdownCPU uint64
	)

	// Iterate over the memory settings
	for ndbinfoDiskWriteSpeedAggregateRows.Next() {
		if err := ndbinfoDiskWriteSpeedAggregateRows.Scan(
			&nodeID, &threadNO, &lcpWrite, &redoWrite, &slowdownIO, &slowdownCPU); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateLcpDesc, prometheus.GaugeValue, float64(lcpWrite),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateRedoDesc, prometheus.GaugeValue, float64(redoWrite),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateIOSlowdownDesc, prometheus.CounterValue, float64(slowdownIO),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskWriteSpeedAggregateCPUSlowdownDesc, prometheus.CounterValue, float64(slowdownCPU),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
	}
	return nil
}
