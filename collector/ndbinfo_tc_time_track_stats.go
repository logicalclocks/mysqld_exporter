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

// Scrape `ndbinfo.tc_time_track_stats`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoTcTimeTrackQuery = `
        SELECT node_id, upper_bound, sum(scans), sum(transactions), sum(read_key_ops),
        sum(write_key_ops), sum(index_key_ops)
        FROM ndbinfo.tc_time_track_stats
        GROUP BY node_id, upper_bound;
	`

var (
	ndbinfoTcTimeTrackScansDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "tc_time_track_scans"),
		"Time track of scans",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoTcTimeTrackTransactionsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "tc_time_track_transactions"),
		"Time track of transactions",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoTcTimeTrackReadKeyDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "tc_time_track_read_key"),
		"Time track of read key operations",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoTcTimeTrackWriteKeyDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "tc_time_track_write_key"),
		"Time track of write key operations",
		[]string{"nodeID", "upperBound"}, nil,
	)
	ndbinfoTcTimeTrackIndexKeyDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "tc_time_track_index_key"),
		"Time track of index key operations",
		[]string{"nodeID", "upperBound"}, nil,
	)
)

// ScrapeNdbinfoTcTimeTrack collects for `ndbinfo.tc_time_track_stats`
type ScrapeNdbinfoTcTimeTrack struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoTcTimeTrack) Name() string {
	return "ndbinfo.tc_time_track_stats"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoTcTimeTrack) Help() string {
	return "Collect metrics from ndbinfo.tc_time_track_stats"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoTcTimeTrack) Version() float64 {
	return 5.7
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoTcTimeTrack) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoTcTimeTrackRows, err := db.QueryContext(ctx, ndbinfoTcTimeTrackQuery)
	if err != nil {
		return err
	}
	defer ndbinfoTcTimeTrackRows.Close()

	var (
		nodeID, upper_bound, scans          uint64
                transactions, read_key_ops          uint64
                write_key_ops, index_key_ops        uint64
	)

	// Iterate over the memory settings
	for ndbinfoTcTimeTrackRows.Next() {
		if err := ndbinfoTcTimeTrackRows.Scan(
			&nodeID, &upper_bound, &scans, &transactions,
                        &read_key_ops, &write_key_ops, &index_key_ops); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoTcTimeTrackScansDesc, prometheus.GaugeValue, float64(scans),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoTcTimeTrackTransactionsDesc, prometheus.GaugeValue, float64(transactions),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoTcTimeTrackReadKeyDesc, prometheus.GaugeValue, float64(read_key_ops),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoTcTimeTrackWriteKeyDesc, prometheus.GaugeValue, float64(write_key_ops),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoTcTimeTrackIndexKeyDesc, prometheus.GaugeValue, float64(index_key_ops),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(upper_bound, 10))

	}
	return nil
}
