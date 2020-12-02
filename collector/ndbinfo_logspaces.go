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

// Scrape `ndbinfo.logspaces`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoLogspacesQuery = `
	SELECT node_id, log_type, log_part, total, used 
	FROM ndbinfo.logspaces;
	`

var (
	ndbinfoLogspacesUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "logspaces_used"),
		"Space used by each log",
		[]string{"nodeID", "logType", "logPart"}, nil,
	)

	ndbinfoLogspacesTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "logspaces_total"),
		"Total space available for each log",
		[]string{"nodeID", "logType", "logPart"}, nil,
	)
)

// ScrapeNdbinfoLogspaces collects for `ndbinfo.logspaces`
type ScrapeNdbinfoLogspaces struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoLogspaces) Name() string {
	return "ndbinfo.logspaces"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoLogspaces) Help() string {
	return "Collect metrics from ndbinfo.logspaces"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoLogspaces) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoLogspaces) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoLogspacesRows, err := db.QueryContext(ctx, ndbinfoLogspacesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoLogspacesRows.Close()

	var (
		nodeID, logPart, used, total        uint64
		logType                             string
	)

	// Iterate over the memory settings
	for ndbinfoLogspacesRows.Next() {
		if err := ndbinfoLogspacesRows.Scan(
			&nodeID, &logType, &logPart, &total, &used); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogspacesUsedDesc, prometheus.GaugeValue, float64(used),
			strconv.FormatUint(nodeID, 10), logType, strconv.FormatUint(logPart, 10))

		ch <- prometheus.MustNewConstMetric(
			ndbinfoLogspacesTotalDesc, prometheus.GaugeValue, float64(total),
			strconv.FormatUint(nodeID, 10), logType, strconv.FormatUint(logPart, 10))
	}
	return nil
}
