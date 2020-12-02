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

// Scrape `ndbinfo.diskpagebuffers`

package collector

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const ndbinfoDiskpagebuffersQuery = `
    SELECT node_id, block_instance, pages_written,
    	   pages_written_lcp, pages_read, log_waits,
    	   page_requests_direct_return, page_requests_wait_queue,
    	   page_requests_wait_io
    FROM ndbinfo.diskpagebuffer;
	`

var (
	ndbinfoDiskpagebuffersPagesWrittenDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_pages_written"),
		"Number of pages written to disk for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskpagebuffersPagesWrittenLcpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_pages_written_lcp"),
		"Number of pages written by local checkpoints for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskpagebuffersPagesReadDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_pages_read"),
		"Number of pages read from disk for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskpagebuffersLogWaitsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_log_waits"),
		"Number of page writes waiting for log to be written to disk for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskpagebuffersDirectDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_direct_return"),
		"Number of requests for pages that were available in buffer for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskpagebuffersQueueDesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_wait_queue"),
		"Number of requests that had to wait for pages to become available in buffer for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
	ndbinfoDiskpagebuffersIODesc = prometheus.NewDesc(
		prometheus.BuildFQName("ndb", ndbinfo, "diskpagebuffer_wait_io"),
		"Number of requests that had to be read from disk for each node, block and thread",
		[]string{"nodeID", "threadNO"}, nil,
	)
)

// ScrapeNdbinfoDiskpagebuffers collects for `ndbinfo.diskpagebuffers`
type ScrapeNdbinfoDiskpagebuffers struct{}

// Name of the Scraper. Should be unique.
func (ScrapeNdbinfoDiskpagebuffers) Name() string {
	return "ndbinfo.diskpagebuffers"
}

// Help describes the role of the Scraper
func (ScrapeNdbinfoDiskpagebuffers) Help() string {
	return "Collect metrics from ndbinfo.diskpagebuffers"
}

// Version of MySQL from which scraper is available
func (ScrapeNdbinfoDiskpagebuffers) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric
func (ScrapeNdbinfoDiskpagebuffers) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	ndbinfoDiskpagebuffersRows, err := db.QueryContext(ctx, ndbinfoDiskpagebuffersQuery)
	if err != nil {
		return err
	}
	defer ndbinfoDiskpagebuffersRows.Close()

	var (
		nodeID, threadNO, pagesWritten                                      uint64
		pagesWrittenLcp, pagesRead, logWaits                                uint64
		pageRequestsDirectReturn, pageRequestsWaitQueue, pageRequestsWaitIO uint64
	)

	// Iterate over the memory settings
	for ndbinfoDiskpagebuffersRows.Next() {
		if err := ndbinfoDiskpagebuffersRows.Scan(
			&nodeID, &threadNO, &pagesWritten, &pagesWrittenLcp,
			&pagesRead, &logWaits, &pageRequestsDirectReturn,
			&pageRequestsWaitQueue, &pageRequestsWaitIO); err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersPagesWrittenDesc, prometheus.CounterValue, float64(pagesWritten),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersPagesWrittenLcpDesc, prometheus.CounterValue, float64(pagesWrittenLcp),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersPagesReadDesc, prometheus.CounterValue, float64(pagesRead),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersLogWaitsDesc, prometheus.CounterValue, float64(logWaits),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersDirectDesc, prometheus.CounterValue, float64(pageRequestsDirectReturn),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersQueueDesc, prometheus.CounterValue, float64(pageRequestsWaitQueue),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
		ch <- prometheus.MustNewConstMetric(
			ndbinfoDiskpagebuffersIODesc, prometheus.CounterValue, float64(pageRequestsWaitIO),
			strconv.FormatUint(nodeID, 10), strconv.FormatUint(threadNO, 10))
	}
	return nil
}
