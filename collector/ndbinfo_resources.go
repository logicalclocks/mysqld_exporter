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

const ndbinfoResourcesQuery = `
	SELECT node_id, resource_name, reserved, used
	FROM ndbinfo.resources;
	`

const ndbinfoLongSignalQuery = `
	SELECT node_id, used_pages, total_pages 
	FROM ndbinfo.resources
        WHERE memory_type = "Long message buffer";
	`

var (
	ndbinfoResourcesReservedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "memory_resource_reserved"),
		"Memory used for each node and memory type in bytes",
		[]string{"nodeID", "memoryType"}, nil,
	)

	ndbinfoResourcesUsedDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "memory_resource_used"),
		"Total memory configured for each node and memory type in bytes",
		[]string{"nodeID", "memoryType"}, nil,
	)

)

// ScrapeNdbinfoResources collects for `ndbinfo.resources`
type ScrapeNdbinfoResources struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoResources) Name() string {
	return "ndbinfo.resources"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoResources) Help() string {
	return "Collect metrics from ndbinfo.resources"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoResources) Version() float64 {
	return 5.7
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoResources) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoResourcesRows, err := db.QueryContext(ctx, ndbinfoResourcesQuery)
	if err != nil {
		return err
	}
	defer ndbinfoResourcesRows.Close()

	var (
		nodeID, used, reserved                     uint64
                used_bytes, reserved_bytes, total_bytes    uint64
                used_pages, total_pages                    uint64
		memoryType                                 string
	)

	// Iterate over the memory settings
	for ndbinfoResourcesRows.Next() {
		if err := ndbinfoResourcesRows.Scan(
			&nodeID, &memoryType, &reserved, &used); err != nil {
			return err
		}
                reserved_bytes = reserved * 32768;
                used_bytes = used * 32768;
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesReservedDesc, prometheus.GaugeValue, float64(reserved_bytes),
			strconv.FormatUint(nodeID, 10), memoryType)
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesUsedDesc, prometheus.GaugeValue, float64(used_bytes),
			strconv.FormatUint(nodeID, 10), memoryType)
	}

	ndbinfoLongSignalMemoryRows, err_long := db.QueryContext(ctx, ndbinfoLongSignalMemoryQuery)
	if err_long != nil {
		return err_long
	}
	defer ndbinfoLongSignalMemoryRows.Close()

	for ndbinfoLongSignalMemoryRows.Next() {
		if err := ndbinfoLongSignalMemoryRows.Scan(
			&nodeID, &used_pages, &total_pages); err != nil {
			return err
		}
                // Convert to bytes from pages
                total_bytes = total_pages * 256
                used_bytes = used_pages * 256
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesReservedDesc, prometheus.GaugeValue, float64(total_bytes),
			strconv.FormatUint(nodeID, 10), "LONG_SIGNAL_MEMORY")
		ch <- prometheus.MustNewConstMetric(
			ndbinfoResourcesUsedDesc, prometheus.GaugeValue, float64(used_bytes),
			strconv.FormatUint(nodeID, 10), "LONG_SIGNAL_MEMORY")
	}
	return nil
}
