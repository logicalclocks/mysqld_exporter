// Copyright 2019, 2020 The Prometheus Authors, LogicalClocks AB
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

// Scrape `ndbinfo.resources`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoFreeMemoryQuery = `
	SELECT node_id, resource_name, reserved, used
	FROM ndbinfo.resources WHERE
        resource_name = "TRANSACTION_MEMORY" OR
        resource_name = "JOBBUFFER" OR
        resource_name = "DATA_MEMORY" OR
        resource_name = "TOTAL_GLOBAL_MEMORY" OR
        resource_name = "SCHEMA_TRANS_MEMORY";
	`

const ndbinfoLongSignalMemoryQuery = `
	SELECT node_id, used_pages, total_pages
	FROM ndbinfo.memoryusage WHERE
        memory_type = "Long message buffer";
	`

var (
	ndbinfoFreeMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "free_memory"),
		"Memory free for each node and memory type in bytes",
		[]string{"nodeID", "memoryType"}, nil,
	)

)

// ScrapeNdbinfoFreeMemory collects for ndbinfo.resources
type ScrapeNdbinfoFreeMemory struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoFreeMemory) Name() string {
	return "ndbinfo.resources.free"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoFreeMemory) Help() string {
	return "Collect metrics from ndbinfo.resources.free"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoFreeMemory) Version() float64 {
	return 5.7
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoFreeMemory) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoFreeMemoryRows, err := db.QueryContext(ctx, ndbinfoFreeMemoryQuery)
	if err != nil {
		return err
	}
	defer ndbinfoFreeMemoryRows.Close()

	var (
		nodeID, reserved, used, free               uint64
                used_pages, total_pages                    uint64
		memoryType                                 string
                free64                                     float64
	)

	for ndbinfoFreeMemoryRows.Next() {
		if err := ndbinfoFreeMemoryRows.Scan(
			&nodeID, &memoryType, &reserved, &used); err != nil {
			return err
		}
                if reserved >= used {
                  free = reserved - used
                } else {
                  free = 0
                }
                // Convert from pages to bytes
                free64 = float64(free)
                free64 = free64 * float64(32768)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFreeMemoryDesc, prometheus.GaugeValue, free64,
			strconv.FormatUint(nodeID, 10), memoryType)
	}

	ndbinfoLongSignalMemoryRows, err := db.QueryContext(ctx, ndbinfoLongSignalMemoryQuery)
	if err != nil {
		return err
	}
	defer ndbinfoLongSignalMemoryRows.Close()

	for ndbinfoLongSignalMemoryRows.Next() {
		if err := ndbinfoLongSignalMemoryRows.Scan(
			&nodeID, &used_pages, &total_pages); err != nil {
			return err
		}
                if total_pages >= used_pages {
                  free = total_pages - used_pages
                } else {
                  free = 0
                }
                // Convert from pages to bytes
                free64 = float64(free)
                free64 = free64 * float64(256)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoFreeMemoryDesc, prometheus.GaugeValue, free64,
			strconv.FormatUint(nodeID, 10), "LONG_SIGNAL_MEMORY")
	}
	return nil
}
