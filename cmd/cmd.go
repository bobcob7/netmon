package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bobcob7/netmon/pkg/stats"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
)

var interfaceColors = []cell.Color{
	cell.Color(2), // Red
	cell.Color(3), // Lime
	cell.Color(4), // Yellow
	cell.Color(5), // Blue
	cell.Color(6), // Fuchsia
	cell.Color(7), // Aqua
}

var brightInterfaceColors = []cell.Color{
	cell.Color(10), // Red
	cell.Color(11), // Lime
	cell.Color(12), // Yellow
	cell.Color(13), // Blue
	cell.Color(14), // Fuchsia
	cell.Color(15), // Aqua
}

func Run(interfaces []string) {
	t, err := termbox.New(
		termbox.ColorMode(terminalapi.ColorMode256),
	)
	if err != nil {
		panic(err)
	}
	defer t.Close()
	ctx, cancel := context.WithCancel(context.Background())

	builder := grid.New()
	tables := make([]*text.Text, len(interfaces))
	interfaceStats := make([]stats.InterfaceStats, len(interfaces))
	tableWidgets := make([]grid.Element, len(interfaces), len(interfaces)+1)
	packetLegend, _ := text.New()
	byteLegend, _ := text.New()
	errorsLegend, _ := text.New()
	droppedLegend, _ := text.New()
	for i, interfaceName := range interfaces {
		interfaceStats[i], _ = stats.NewInterfaceStats(interfaceName)
		tables[i], _ = text.New(
			text.DisableScrolling(),
		)
		tables[i].Write(interfaceStats[i].Write())
		color := interfaceColors[i]

		packetLegend.Write(fmt.Sprintf("TX Packets: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(interfaceColors[i])))
		packetLegend.Write(fmt.Sprintf("RX Packets: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(brightInterfaceColors[i])))
		byteLegend.Write(fmt.Sprintf("TX Bytes: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(interfaceColors[i])))
		byteLegend.Write(fmt.Sprintf("RX Bytes: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(brightInterfaceColors[i])))
		errorsLegend.Write(fmt.Sprintf("TX Errors: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(interfaceColors[i])))
		errorsLegend.Write(fmt.Sprintf("RX Errors: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(brightInterfaceColors[i])))
		droppedLegend.Write(fmt.Sprintf("TX Dropped: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(interfaceColors[i])))
		droppedLegend.Write(fmt.Sprintf("RX Dropped: %s\n", interfaceName), text.WriteCellOpts(cell.FgColor(brightInterfaceColors[i])))

		tableWidgets[i] = grid.RowHeightFixed(9,
			grid.Widget(tables[i],
				container.Border(linestyle.Light),
				container.BorderColor(color),
			),
		)
	}
	legendWidget := grid.RowHeightFixed(9,
		grid.Widget(packetLegend,
			container.Border(linestyle.Double),
			container.ID("legend"),
		),
	)
	tableWidgets = append(tableWidgets, legendWidget)
	builder.Add(
		grid.ColWidthFixed(42, tableWidgets...),
	)

	packetGraph, err := linechart.New(
		linechart.YAxisFormattedValues(linechart.ValueFormatterRound),
	)
	if err != nil {
		panic(err)
	}
	byteGraph, err := linechart.New(
		linechart.YAxisFormattedValues(linechart.ValueFormatterRound),
	)
	if err != nil {
		panic(err)
	}
	errorsGraph, err := linechart.New(
		linechart.YAxisFormattedValues(linechart.ValueFormatterRound),
	)
	if err != nil {
		panic(err)
	}
	droppedGraph, err := linechart.New(
		linechart.YAxisFormattedValues(linechart.ValueFormatterRound),
	)
	if err != nil {
		panic(err)
	}

	builder.Add(
		grid.ColWidthPerc(10, grid.Widget(
			packetGraph,
			container.Border(linestyle.Light), container.BorderTitle("P=pkt/s B=byte/s E=Errors D=Dropped R=Reset Q=Quit"),
			container.ID("graph"),
		)),
	)
	gridWidget, err := builder.Build()
	if err != nil {
		panic(err)
	}

	c, err := container.New(t, gridWidget...)
	if err != nil {
		panic(err)
	}

	currentSelection := 'P'
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
		if k.Key == 'r' || k.Key == 'R' {
			stats.Reset()
		}
		if (k.Key == 'p' || k.Key == 'P') && currentSelection != 'P' {
			currentSelection = 'P'
			if err := c.Update("graph", container.PlaceWidget(packetGraph)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
			if err := c.Update("legend", container.PlaceWidget(packetLegend)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
		}
		if (k.Key == 'b' || k.Key == 'B') && currentSelection != 'B' {
			currentSelection = 'B'
			if err := c.Update("graph", container.PlaceWidget(byteGraph)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
			if err := c.Update("legend", container.PlaceWidget(byteLegend)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
		}
		if (k.Key == 'e' || k.Key == 'E') && currentSelection != 'E' {
			currentSelection = 'E'
			if err := c.Update("graph", container.PlaceWidget(errorsGraph)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
			if err := c.Update("legend", container.PlaceWidget(errorsLegend)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
		}
		if (k.Key == 'd' || k.Key == 'D') && currentSelection != 'D' {
			currentSelection = 'D'
			if err := c.Update("graph", container.PlaceWidget(droppedGraph)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
			if err := c.Update("legend", container.PlaceWidget(droppedLegend)); err != nil {
				log.Fatalf("c.Update => %v", err)
			}
		}
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			<-ticker.C
			for i, table := range tables {
				inter, _ := stats.NewInterfaceStats(interfaces[i])
				table.Write(inter.Write(), text.WriteReplace())
				brightColor := brightInterfaceColors[i]
				color := interfaceColors[i]
				series := inter.Graph()
				packetGraph.Series(
					fmt.Sprintf("TX Packets: %s", interfaces[i]),
					series.TxPackets,
					linechart.SeriesCellOpts(cell.FgColor(brightColor)),
				)
				packetGraph.Series(
					fmt.Sprintf("RX Packets: %s", interfaces[i]),
					series.RxPackets,
					linechart.SeriesCellOpts(cell.FgColor(color)),
				)
				byteGraph.Series(
					fmt.Sprintf("TX Bytes: %s", interfaces[i]),
					series.TxBytes,
					linechart.SeriesCellOpts(cell.FgColor(brightColor)),
				)
				byteGraph.Series(
					fmt.Sprintf("RX Bytes: %s", interfaces[i]),
					series.RxBytes,
					linechart.SeriesCellOpts(cell.FgColor(color)),
				)
				byteGraph.Series(
					fmt.Sprintf("TX Errors: %s", interfaces[i]),
					series.TxErrors,
					linechart.SeriesCellOpts(cell.FgColor(brightColor)),
				)
				byteGraph.Series(
					fmt.Sprintf("RX Errors: %s", interfaces[i]),
					series.RxErrors,
					linechart.SeriesCellOpts(cell.FgColor(color)),
				)
				byteGraph.Series(
					fmt.Sprintf("TX Dropped: %s", interfaces[i]),
					series.TxDropped,
					linechart.SeriesCellOpts(cell.FgColor(brightColor)),
				)
				byteGraph.Series(
					fmt.Sprintf("RX Dropped: %s", interfaces[i]),
					series.RxDropped,
					linechart.SeriesCellOpts(cell.FgColor(color)),
				)
			}
		}
	}()

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
