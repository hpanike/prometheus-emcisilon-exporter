/*
Copyright 2018 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package collector

import (
	"fmt"

	"github.com/adobe/prometheus-emcisilon-exporter/isiclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type diskCollector struct {
	diskBusy            *prometheus.Desc
	diskIoschedQueueAll *prometheus.Desc
	diskXfersInRateAll  *prometheus.Desc
	diskXfersOutRateAll *prometheus.Desc
}

func init() {
	registerCollector("disk", defaultEnabled, NewDiskCollector)
}

//NewDiskCollector returns a new Collector exposing node disk statistics.
func NewDiskCollector() (Collector, error) {
	return &diskCollector{
		diskBusy: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "disk_busy_all"),
			"Current disk busy percentage represented in 0.0-1.0.",
			[]string{"node", "disk"}, ConstLabels,
		),
		diskIoschedQueueAll: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "disk_iosched_queued_all"),
			"Current queue depth for IO sceduler.",
			[]string{"node", "disk"}, ConstLabels,
		),
		diskXfersInRateAll: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "disk_xfers_in_rate_all"),
			"Current disk ingest transfer rate.",
			[]string{"node", "disk"}, ConstLabels,
		),
		diskXfersOutRateAll: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nodeCollectorSubsystem, "disk_xfers_out_rate_all"),
			"Current disk egress transfer rate.",
			[]string{"node", "disk"}, ConstLabels,
		),
	}, nil
}

func (c *diskCollector) Update(ch chan<- prometheus.Metric) error {
	keyMap := make(map[string]string)

	keyMap["diskBusy"] = "node.disk.busy.all"
	keyMap["diskIoschedQueueAll"] = "node.disk.iosched.queue.all"
	keyMap["diskXfersInRateAll"] = "node.disk.xfers.in.rate.all"
	keyMap["diskXfersOutRateAll"] = "node.disk.xfers.out.rate.all"

	for promStat, statKey := range keyMap {
		resp, err := isiclient.QueryStatsEngineMultiVal(IsiCluster.Client, statKey)
		if err != nil {
			log.Warnf("Error attempting to query stats engine with key %s: %s", statKey, err)
			return err
		}
		for _, stat := range resp.Stats {
			node := fmt.Sprintf("%v", stat.Devid)
			for _, valset := range stat.ValueSet {
				for disk, val := range valset {
					c.updateDiskStats(promStat, node, disk, val, ch)
				}
			}
		}
	}
	return nil
}

func (c *diskCollector) updateDiskStats(promStat string, node string, disk string, val float64, ch chan<- prometheus.Metric) error {
	switch promStat {
	case "diskBusy":
		mVal := val / 10
		ch <- prometheus.MustNewConstMetric(c.diskBusy, prometheus.GaugeValue, mVal, node, disk)
	case "diskIoschedQueueAll":
		ch <- prometheus.MustNewConstMetric(c.diskIoschedQueueAll, prometheus.GaugeValue, val, node, disk)
	case "diskXfersInRateAll":
		ch <- prometheus.MustNewConstMetric(c.diskXfersInRateAll, prometheus.GaugeValue, val, node, disk)
	case "diskXfersOutRateAll":
		ch <- prometheus.MustNewConstMetric(c.diskXfersOutRateAll, prometheus.GaugeValue, val, node, disk)
	}
	return nil
}
