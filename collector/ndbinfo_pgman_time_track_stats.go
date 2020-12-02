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

// Scrape `ndbinfo.pgman_time_track_stats`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoPgmanTimeTrackQuery = `
        SELECT node_id, upper_bound, sum(page_reads), sum(page_writes),
        sum(log_waits), sum(get_page)
        FROM ndbinfo.pgman_time_track_stats
        GROUP BY node_id, upper_bound;
	`

var (
	ndbinfoPgmanTimeTrackPageReadsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "pgman_time_track_page_reads"),
		"Time track of page reads",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoPgmanTimeTrackPageWritesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "pgman_time_track_page_writes"),
		"Time track of page writes",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoPgmanTimeTrackLogWaitsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "pgman_time_track_log_waits"),
		"Time track of wait for UNDO log writes",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoPgmanTimeTrackGetPageDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "pgman_time_track_get_page"),
		"Time track of get_page operation",
		[]string{"nodeID", "upperBound"}, nil,
	)
)

// ScrapeNdbinfoPgmanTimeTrack collects for `ndbinfo.pgman_time_track_stats`
type ScrapeNdbinfoPgmanTimeTrack struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoPgmanTimeTrack) Name() string {
	return "ndbinfo.pgman_time_track_stats"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoPgmanTimeTrack) Help() string {
	return "Collect metrics from ndbinfo.pgman_time_track_stats"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoPgmanTimeTrack) Version() float64 {
	return 5.7
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoPgmanTimeTrack) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoPgmanTimeTrackRows, err := db.QueryContext(ctx, ndbinfoPgmanTimeTrackQuery)
	if err != nil {
		return err
	}
	defer ndbinfoPgmanTimeTrackRows.Close()

	var (
		nodeID, upper_bound, page_reads     uint64
                page_writes, log_waits              uint64
                get_page                            float64
	)

	// Iterate over the memory settings
	for ndbinfoPgmanTimeTrackRows.Next() {
		if err := ndbinfoPgmanTimeTrackRows.Scan(
			&nodeID, &upper_bound, &page_reads, &page_writes,
                        &log_waits, &get_page); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoPgmanTimeTrackPageReadsDesc, prometheus.GaugeValue, float64(page_reads),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoPgmanTimeTrackPageWritesDesc, prometheus.GaugeValue, float64(page_writes),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoPgmanTimeTrackLogWaitsDesc, prometheus.GaugeValue, float64(log_waits),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoPgmanTimeTrackGetPageDesc, prometheus.GaugeValue, float64(get_page),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))
	}
	return nil
}
