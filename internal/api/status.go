// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

func styleFromName(style string) (table.Style, bool) {
	switch style {
	case "default":
		return table.StyleDefault, true
	case "bold":
		return table.StyleBold, true
	case "rounded":
		return table.StyleRounded, true
	case "light":
		return table.StyleLight, true
	case "double":
		return table.StyleDouble, true
	case "colored-bright":
		return table.StyleColoredBright, true
	case "colored-dark":
		return table.StyleColoredDark, true
	}

	return table.Style{}, false
}

func renderTable(t table.Writer, f io.Writer, format string) {
	t.SetOutputMirror(f)

	var s table.Style
	if strings.HasPrefix(format, "pretty-") {
		styleName := strings.TrimPrefix(format, "pretty-")
		format = "pretty"

		if ss, ok := styleFromName(styleName); ok {
			s = ss
		} else {
			s = table.StyleDefault
		}
	} else {
		s = table.StyleDefault
	}

	s.Format.Header = text.FormatDefault

	t.SetStyle(s)

	switch format {
	case "pretty":
		t.Render()
	case "csv":
		t.RenderCSV()
	case "html":
		t.RenderHTML()
	case "md", "markdown":
		t.RenderMarkdown()
	case "tsv":
		t.RenderTSV()
	}
}

func (s *Status) Print(f io.Writer, format string) {
	if format == "json" {
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		enc.Encode(s)

		return
	}

	s.PrintBreakers(f, format)

	if len(s.Groups) > 0 {
		fmt.Fprintln(f)
		s.PrintGroups(f, format)
	}

	if len(s.Outlets) > 0 {
		fmt.Fprintln(f)
		s.PrintOutlets(f, format)
	}
}

func (s *Status) PrintBreakers(f io.Writer, format string) {
	t := table.NewWriter()
	t.AppendHeader(table.Row{
		"Circuit",
		"True RMS Current",
		"Peak RMS Current",
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
	})

	for _, breaker := range s.Breakers {
		t.AppendRow(table.Row{
			breaker.Name,
			withUnit(breaker.TrueRMSCurrent, "A", 1),
			withUnit(breaker.PeakRMSCurrent, "A", 1),
		})
	}

	renderTable(t, f, format)
}

func (s *Status) PrintGroups(f io.Writer, format string) {
	t := table.NewWriter()
	t.AppendHeader(table.Row{
		"Group",
		"True RMS Current",
		"Peak RMS Current",
		"True RMS Voltage",
		"Average Power",
		"Power",
		"Energy",
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
	})

	for _, group := range s.Groups {
		t.AppendRow(table.Row{
			group.Name,
			withUnit(group.TrueRMSCurrent, "A", 1),
			withUnit(group.PeakRMSCurrent, "A", 1),
			withUnit(group.TrueRMSVoltage, "V", 1),
			withUnit(group.AveragePower, "W", 1),
			withUnit(group.Power, "W", 1),
			withUnit(group.Energy, "kWh", 3),
		})
	}

	renderTable(t, f, format)
}

func (s *Status) PrintOutlets(f io.Writer, format string) {
	t := table.NewWriter()
	t.AppendHeader(table.Row{
		"Outlet",
		"True RMS Current",
		"Peak RMS Current",
		"True RMS Voltage",
		"Average Power",
		"Power",
		"Energy",
		"State",
		"Locked",
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
	})

	for _, outlet := range s.Outlets {
		t.AppendRow(table.Row{
			outlet.Name,
			withUnit(outlet.TrueRMSCurrent, "A", 1),
			withUnit(outlet.PeakRMSCurrent, "A", 1),
			withUnit(outlet.TrueRMSVoltage, "V", 1),
			withUnit(outlet.AveragePower, "W", 1),
			withUnit(outlet.Power, "W", 1),
			withUnit(outlet.Energy, "kWh", 3),
			outlet.State,
			outlet.Locked,
		})
	}

	renderTable(t, f, format)
}

func (s *OutletStatus) Print(f io.Writer, format string) {
	if format == "json" {
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		enc.Encode(s)

		return
	}

	t := table.NewWriter()
	t.AppendHeader(table.Row{
		"Outlet",
		"True RMS Current",
		"Peak RMS Current",
		"True RMS Voltage",
		"Average Power",
		"Power",
		"Energy",
		"State",
		"Locked",
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
	})
	t.AppendRow(table.Row{
		s.Name,
		withUnit(s.TrueRMSCurrent, "A", 1),
		withUnit(s.PeakRMSCurrent, "A", 1),
		withUnit(s.TrueRMSVoltage, "V", 1),
		withUnit(s.AveragePower, "W", 1),
		withUnit(s.Power, "W", 1),
		withUnit(s.Energy, "kWh", 3),
		s.State,
		s.Locked,
	})

	renderTable(t, f, format)
}

func withUnit(n float32, unit string, digits int) string {
	fmt := message.NewPrinter(language.English)
	return fmt.Sprintf("%v %s", number.Decimal(n, number.MinFractionDigits(digits), number.MaxFractionDigits(digits)), unit)
}
