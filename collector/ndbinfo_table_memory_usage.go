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

// Scrape `ndbinfo.table_memory_usage`

package collector

import (
	"context"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoTableMemoryUsageQuery = `
	SELECT database_name, table_name, in_memory_bytes, free_in_memory_bytes, disk_memory_bytes, free_disk_memory_bytes
	FROM ndbinfo.table_memory_usage;
	`

var (
	ndbTableMemoryInMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "table_memory_bytes"),
		"Bytes of memory currently used in memory for the table",
		[]string{"database", "table"}, nil,
	)

	ndbTableMemoryFreeInMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "table_free_memory_bytes"),
		"Bytes of free memory in memory for the table",
		[]string{"database", "table"}, nil,
	)

	ndbTableMemoryDiskMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "table_disk_bytes"),
		"Bytes of disk memory used by the table",
		[]string{"database", "table"}, nil,
	)

	ndbTableMemoryFreeDiskMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "table_free_disk_bytes"),
		"Bytes of free disk memory for the table",
		[]string{"database", "table"}, nil,
	)
)

type ScrapeNdbinfoTableMemoryUsage struct{}

func (ScrapeNdbinfoTableMemoryUsage) Name() string {
	return "ndbinfo.table_memory_usage"
}

func (ScrapeNdbinfoTableMemoryUsage) Help() string {
	return "Collect metrics from ndbinfo.table_memory_usage"
}

func (ScrapeNdbinfoTableMemoryUsage) Version() float64 {
	return 5.7
}

func (ScrapeNdbinfoTableMemoryUsage) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx, ndbinfoTableMemoryUsageQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var (
		databaseName        string
		tableName           string
		inMemoryBytes       uint64
		freeInMemoryBytes   uint64
		diskMemoryBytes     uint64
		freeDiskMemoryBytes uint64
	)

	for rows.Next() {
		if err := rows.Scan(
			&databaseName,
			&tableName,
			&inMemoryBytes,
			&freeInMemoryBytes,
			&diskMemoryBytes,
			&freeDiskMemoryBytes,
		); err != nil {
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			ndbTableMemoryInMemoryDesc,
			prometheus.GaugeValue,
			float64(inMemoryBytes),
			databaseName,
			tableName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbTableMemoryFreeInMemoryDesc,
			prometheus.GaugeValue,
			float64(freeInMemoryBytes),
			databaseName,
			tableName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbTableMemoryDiskMemoryDesc,
			prometheus.GaugeValue,
			float64(diskMemoryBytes),
			databaseName,
			tableName,
		)
		ch <- prometheus.MustNewConstMetric(
			ndbTableMemoryFreeDiskMemoryDesc,
			prometheus.GaugeValue,
			float64(freeDiskMemoryBytes),
			databaseName,
			tableName,
		)
	}

	return nil
}