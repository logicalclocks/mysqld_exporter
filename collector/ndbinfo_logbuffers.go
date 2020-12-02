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

// Scrape `ndbinfo.logbuffers`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoLogbuffersQuery = `
	SELECT node_id, log_type, log_part, total, used 
	FROM ndbinfo.logbuffers;
	`

var (
	ndbinfoLogbuffersUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "logbuffers_used"),
		"Buffer space used by each log",
		[]string{"nodeID", "logType", "logPart"}, nil,
	)

	ndbinfoLogbuffersTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "logbuffers_total"),
		"Total buffer space available for each log",
		[]string{"nodeID", "logType", "logPart"}, nil,
	)
)

// ScrapeNdbinfoLogbuffers collects for `ndbinfo.logbuffers`
type ScrapeNdbinfoLogbuffers struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoLogbuffers) Name() string {
	return "ndbinfo.logbuffers"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoLogbuffers) Help() string {
	return "Collect metrics from ndbinfo.logbuffers"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoLogbuffers) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoLogbuffers) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoLogbuffersRows, err := db.QueryContext(ctx, ndbinfoLogbuffersQuery)
	if err != nil {
		return err
	}
	defer ndbinfoLogbuffersRows.Close()

	var (
		nodeID, logPart, used, total        uint64
		logType                             string
	)

	// Iterate over the memory settings
	for ndbinfoLogbuffersRows.Next() {
		if err := ndbinfoLogbuffersRows.Scan(
			&nodeID, &logType, &logPart, &total, &used); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogbuffersUsedDesc, prometheus.GaugeValue, float64(used),
			strconv.FormatUint(nodeID, 10), logType, strconv.FormatUint(logPart, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogbuffersTotalDesc, prometheus.GaugeValue, float64(total),
			strconv.FormatUint(nodeID, 10), logType, strconv.FormatUint(logPart, 10))
	}
	return nil
}
