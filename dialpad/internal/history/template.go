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

func RequestForm(phone string, events []SmsEvent) []byte {
	titlePhone := ""
	if phone != "" {
		titlePhone = " with " + html.EscapeString(phone)
	}
	head := fmt.Sprintf(`
<head>
	<title>SMS History %s</title>
	<meta charset="utf-8" />
	<style>
		body {
			font-family: sans-serif;
		}
		form {
			max-width: 400px;
			margin: 10px auto;
			padding: 20px;
			background-color: #f0f0f0;
			box-shadow: 0px 0px 10px #888888;
			text-align: center;
			align: center;
		}
		.message {
			color: red;
			text-align: center;
		}
		.logout {
			text-align: center;
			margin-top: 10px;
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
`, titlePhone)
	form := fmt.Sprintf(`
<form action="/history" method="GET">
	<label for="phone">Phone:</label>
	<input type="text" id="phone" name="phone" value="%s" placeholder="+15106666687" required><br>
	<button type="submit">Search</button>
</form>`, phone)
	page := `<!DOCTYPE html><html>` + head + `<body>`
	page += form
	if len(events) == 0 {
		if phone != "" {
			page += fmt.Sprintf(`<p class="message">You have no text history with %s</p>`, phone)
		} else {
			page += fmt.Sprintf(`<p class="message">Please specify a phone number</p>`)
		}
	} else {
		page += threadTable(phone, events)
	}
	page += `<p class="logout"><a href="/logout">Logout</a></p>`
	page += `</body></html>`
	return []byte(page)
}

func threadTable(phone string, events []SmsEvent) string {
	tableHdr := fmt.Sprintf(`
<table>
<tr>
	<th style="width:"40%%">You</th>
	<th style="width:"55%%">%s</th>
	<th style="width:"5%%">When</th>
</tr>`, html.EscapeString(phone))
	tableFooter := `</table>`
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
	return tableHdr + tableBody + tableFooter
}

func ServerErrorForm(phone string) []byte {
	head := `
<head>
	<title>Error</title>
	<meta charset="utf-8" />
	<style>
		body {
			font-family: sans-serif;
		}
		.message {
			color: red;
			text-align: center;
		}
		.normal {
			text-align: center;
			margin-top: 10px;
		}
	</style>
</head>
`
	page := `<!DOCTYPE html><html>` + head + `<body>`
	page += fmt.Sprintf(`<h1>Error fetching history for %s</h1>`, html.EscapeString(phone))
	page += `<p class="message">Sorry, an unexpected error occurred fetching your SMS history.</p>`
	page += `<p class="normal">Errors like this are usually temporary.</p>`
	page += fmt.Sprintf(`<p class="normal">To try your query again,
				<a href="/history?phone=%s">click here</a>.</p>`, phone)
	page += `<p class="normal"></p>`
	page += `<p class="normal"><a href="/logout">Logout</a></p>`
	page += `</body></html>`
	return []byte(page)
}
