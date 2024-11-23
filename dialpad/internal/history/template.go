/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"fmt"
	"html"
	"strings"
	"time"
)

var (
	PT *time.Location
)

func init() {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic(err)
	}
	PT = loc
}

func FormatThread(phone string, events []SmsEvent) string {
	head := `
<head>
	<meta charset="utf-8" />
	<style>
		body {
			font-family: sans-serif;
		}
		table {
			border: 1px solid black;
		}
		th, td {
			border: 1px solid black;
			padding-top: 2px;
			padding-bottom: 2px;
			padding-left: 10px;
			padding-right: 10px;
		}
	</style>
</head>
`
	bodyTop := `<body><table>`
	tableHdr := `
<tr>
	<th style="width:40%">You</th>
	<th style="width:55%">Client</th>
	<th style="width:5%">Time</th>
</tr>`
	bodyFoot := `</table></body>`
	var rows []string
	for _, event := range events {
		start := `<tr><td>`
		leftMiddle := `</td><td>`
		rightMiddle := `</td><td style="background-color:#D6EEEE">`
		ts := time.UnixMicro(event.Date).In(PT).Format("1/2/06 3:04PM")
		end := fmt.Sprintf("</td><td style=\"color:grey\">%s</td></tr>", ts)
		var content []string
		if event.Text != "" {
			content = append(content, html.EscapeString(event.Text))
		}
		if event.MmsUrl != "" {
			content = append(content, fmt.Sprintf("<img src=%q />", event.MmsUrl))
		}
		var row string
		for _, c := range content {
			if event.FromPhone == phone {
				row = start + rightMiddle + c + end
			} else {
				row = start + c + leftMiddle + end
			}
		}
		rows = append(rows, row)
	}
	tableBody := strings.Join(rows, "")
	page := `<!DOCTYPE html><html>` + head + bodyTop + tableHdr + tableBody + bodyFoot + `</html>`
	return page
}
