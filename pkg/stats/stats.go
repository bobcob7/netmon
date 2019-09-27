package stats

import (
	"errors"
	"fmt"

	"github.com/vishvananda/netlink"
)

type InterfaceStats struct {
	Name          string
	NetworkStats  NetworkStats
	NetworkSeries NetworkSeries
}

type NetworkStats struct {
	TxPackets uint64
	TxBytes   uint64
	TxErrors  uint64
	TxDropped uint64
	RxPackets uint64
	RxBytes   uint64
	RxErrors  uint64
	RxDropped uint64
}

const maxNetworkSeriesFrame = 300

type NetworkSeries struct {
	TxPackets []float64
	TxBytes   []float64
	TxErrors  []float64
	TxDropped []float64
	RxPackets []float64
	RxBytes   []float64
	RxErrors  []float64
	RxDropped []float64
}

func Reset() {
	statCache = make(map[string]NetworkStats)
	seriesCache = make(map[string]*NetworkSeries)
}

func NewNetworkSeries() *NetworkSeries {
	return &NetworkSeries{
		TxPackets: make([]float64, maxNetworkSeriesFrame),
		TxBytes:   make([]float64, maxNetworkSeriesFrame),
		TxErrors:  make([]float64, maxNetworkSeriesFrame),
		TxDropped: make([]float64, maxNetworkSeriesFrame),
		RxPackets: make([]float64, maxNetworkSeriesFrame),
		RxBytes:   make([]float64, maxNetworkSeriesFrame),
		RxErrors:  make([]float64, maxNetworkSeriesFrame),
		RxDropped: make([]float64, maxNetworkSeriesFrame),
	}
}

func getKernelStats(linkName string) *netlink.LinkStatistics {
	// Grab stats from kernel
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		return nil
	}
	attrs := link.Attrs()
	if attrs == nil {
		return nil
	}
	return attrs.Statistics
}

func getAbsoluteStats(linkName string) *NetworkStats {
	// Grab stats from kernel
	kernel := getKernelStats(linkName)
	if kernel == nil {
		return nil
	}
	return &NetworkStats{
		TxPackets: kernel.TxPackets,
		TxBytes:   kernel.TxBytes,
		TxErrors:  kernel.TxErrors,
		TxDropped: kernel.TxDropped,
		RxPackets: kernel.RxPackets,
		RxBytes:   kernel.RxBytes,
		RxErrors:  kernel.RxErrors,
		RxDropped: kernel.RxDropped,
	}
}

func getRelativeStats(linkName string) *NetworkStats {
	absStats := getAbsoluteStats(linkName)
	if absStats == nil {
		return nil
	}
	// Handle network stats
	var relStats NetworkStats
	if oldStats, found := statCache[linkName]; found {
		relStats = absStats.sub(oldStats)
	} else {
		relStats = NetworkStats{}
		statCache[linkName] = *absStats
	}
	// Handle network series
	if series, found := seriesCache[linkName]; found {
		series.add(relStats)
	} else {
		seriesCache[linkName] = NewNetworkSeries()
	}
	return &relStats
}

func (now NetworkStats) sub(before NetworkStats) NetworkStats {
	return NetworkStats{
		TxPackets: now.TxPackets - before.TxPackets,
		TxBytes:   now.TxBytes - before.TxBytes,
		TxErrors:  now.TxErrors - before.TxErrors,
		TxDropped: now.TxDropped - before.TxDropped,
		RxPackets: now.RxPackets - before.RxPackets,
		RxBytes:   now.RxBytes - before.RxBytes,
		RxErrors:  now.RxErrors - before.RxErrors,
		RxDropped: now.RxDropped - before.RxDropped,
	}
}

func (series *NetworkSeries) add(point NetworkStats) {
	series.TxPackets = append([]float64{float64(point.TxPackets)}, series.TxPackets[:maxNetworkSeriesFrame-1]...)
	series.TxBytes = append([]float64{float64(point.TxBytes)}, series.TxBytes[:maxNetworkSeriesFrame-1]...)
	series.TxErrors = append([]float64{float64(point.TxErrors)}, series.TxErrors[:maxNetworkSeriesFrame-1]...)
	series.TxDropped = append([]float64{float64(point.TxDropped)}, series.TxDropped[:maxNetworkSeriesFrame-1]...)
	series.RxPackets = append([]float64{float64(point.RxPackets)}, series.RxPackets[:maxNetworkSeriesFrame-1]...)
	series.RxBytes = append([]float64{float64(point.RxBytes)}, series.RxBytes[:maxNetworkSeriesFrame-1]...)
	series.RxErrors = append([]float64{float64(point.RxErrors)}, series.RxErrors[:maxNetworkSeriesFrame-1]...)
	series.RxDropped = append([]float64{float64(point.RxDropped)}, series.RxDropped[:maxNetworkSeriesFrame-1]...)
}

func newNetworkStats(now netlink.LinkStatistics, before NetworkStats) NetworkStats {
	return NetworkStats{
		TxPackets: now.TxPackets - before.TxPackets,
		TxBytes:   now.TxBytes - before.TxBytes,
		TxErrors:  now.TxErrors - before.TxErrors,
		TxDropped: now.TxDropped - before.TxDropped,
		RxPackets: now.RxPackets - before.RxPackets,
		RxBytes:   now.RxBytes - before.RxBytes,
		RxErrors:  now.RxErrors - before.RxErrors,
		RxDropped: now.RxDropped - before.RxDropped,
	}
}

var statCache map[string]NetworkStats
var seriesCache map[string]*NetworkSeries

func init() {
	statCache = make(map[string]NetworkStats)
	seriesCache = make(map[string]*NetworkSeries)
}

func NewInterfaceStats(linkName string) (InterfaceStats, error) {
	output := InterfaceStats{
		Name: linkName,
	}
	stats := getRelativeStats(linkName)
	if stats == nil {
		return output, errors.New("Could not get statistics")
	}
	output.NetworkStats = *stats
	return output, nil
}

func (i InterfaceStats) Print() {
	fmt.Printf(`+---------------------------------------+
|Name: %32s |
+---------+-TX-----------+-RX-----------+
|Packets: | %12d | %12d |
|Bytes:   | %12d | %12d |
|Dropped: | %12d | %12d |
|Error:   | %12d | %12d |
+---------+-TX-----------+-RX-----------+
`, i.Name,
		i.NetworkStats.TxPackets, i.NetworkStats.RxPackets,
		i.NetworkStats.TxBytes, i.NetworkStats.RxBytes,
		i.NetworkStats.TxDropped, i.NetworkStats.RxDropped,
		i.NetworkStats.TxErrors, i.NetworkStats.RxErrors,
	)
}

func (i InterfaceStats) Write() string {
	return fmt.Sprintf(` Name: %32s
----------+-TX-----------+-RX----------
 Packets: | %12d | %12d
 Bytes:   | %12d | %12d
 Dropped: | %12d | %12d
 Error:   | %12d | %12d
`, i.Name,
		i.NetworkStats.TxPackets, i.NetworkStats.RxPackets,
		i.NetworkStats.TxBytes, i.NetworkStats.RxBytes,
		i.NetworkStats.TxDropped, i.NetworkStats.RxDropped,
		i.NetworkStats.TxErrors, i.NetworkStats.RxErrors,
	)
}

func (i InterfaceStats) Graph() NetworkSeries {
	if series, found := seriesCache[i.Name]; found {
		return *series
	}
	return NetworkSeries{}
}
