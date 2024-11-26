/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"fmt"
	"html"
	"net/url"
	"slices"
	"strings"
)

func SearchForm(filter string, entries []SearchEntry) []byte {
	escFilter := html.EscapeString(filter)
	titleFilter := ""
	if filter != "" {
		titleFilter = fmt.Sprintf(" filtered by %q", escFilter)
	}
	head := fmt.Sprintf(`
<head>
	<title>Contacts%s</title>
	<meta charset="utf-8" />
	<style>
		body {
			font-family: sans-serif;
		}
		form {
			max-width: 600px;
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
			width: 100%%;
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
`, titleFilter)
	form := fmt.Sprintf(`
<form action="/search" method="GET">
	<label for="filter">Filter:</label>
	<input type="text" id="filter" name="filter" value="%s" placeholder="name or phone" size="30"><br>
	<button type="submit">Filter</button>
</form>`, filter)
	page := `<!DOCTYPE html><html>` + head + `<body>`
	page += form
	if len(entries) == 0 {
		if filter != "" {
			page += fmt.Sprintf(`<p class="message">You have no contacts that match %q</p>`, escFilter)
		} else {
			page += fmt.Sprintf(`<p class="message">Something went wrong. Please reload this page.</p>`)
		}
	} else {
		page += searchTable(entries)
	}
	page += `<p class="logout"><a href="/logout">Logout</a></p>`
	page += `</body></html>`
	return []byte(page)
}

func searchTable(entries []SearchEntry) string {
	slices.SortFunc(entries, SearchEntryCompare)
	columns, width, rows := 4, 25, (len(entries)+3)/4
	var tableRows []string
	for i := 0; i < rows; i++ {
		row := `<tr>`
		for j, k := i, 0; k < columns; j, k = j+rows, k+1 {
			col := fmt.Sprintf(`<td width="%d%%"></td>`, width)
			if j < len(entries) {
				entry := entries[j]
				link := fmt.Sprintf(`/history?phone=%s&name=%s`, url.QueryEscape(entry.Phone), url.QueryEscape(entry.FullName))
				name := fmt.Sprintf(`<a href="%s">%s</a>`, link, html.EscapeString(entry.FullName))
				num := fmt.Sprintf(`<a href="%s">%s</a>`, link, FormatForHTML(entry.Phone))
				col = fmt.Sprintf(`<td width="%d%%">%s<br />%s</td>`, width, name, num)
			}
			row += col
		}
		row += `</tr>`
		tableRows = append(tableRows, row)
	}
	tableBody := strings.Join(tableRows, "")
	return `<table>` + tableBody + `</table>`
}

func ServerErrorForm(filter string) []byte {
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
	page += fmt.Sprintf(`<h1>Error filtering contacts for %s</h1>`, html.EscapeString(filter))
	page += `<p class="message">Sorry, an unexpected error occurred filtering your contacts.</p>`
	page += `<p class="normal">Errors like this are usually temporary.</p>`
	page += fmt.Sprintf(`<p class="normal">To try your query again,
				<a href="/search?filter=%s">click here</a>.</p>`, filter)
	page += `<p class="normal"></p>`
	page += `<p class="normal"><a href="/logout">Logout</a></p>`
	page += `</body></html>`
	return []byte(page)
}

// FormatForHTML takes an E.164 phone number and formats
// it so it displays naturally in HTML.
func FormatForHTML(phone string) string {
	dash := "&#8209;" // non-breaking hyphen
	if strings.HasPrefix(phone, "+1") {
		// Zone 1
		return fmt.Sprintf("(%s)&nbsp;%s%s%s", phone[2:5], phone[5:8], dash, phone[8:12])
	} else {
		// international, separate the country code
		prefix := phone[:3]
		suffix := phone[3:]
		if !twoDigitCountryCodes.Contains(phone[1:3]) {
			prefix = phone[:4]
			suffix = phone[4:]
		}
		var suffixSuffix string
		if len(suffix)%3 == 1 {
			// put last 4 together
			suffix, suffixSuffix = suffix[:len(suffix)-4], dash+suffix[len(suffix)-4:]
		}
		for i := len(suffix); i > 3; i = i - 3 {
			suffix = suffix[0:i] + dash + suffix[i-3:i]
		}
		return prefix + dash + suffix + suffixSuffix
	}
}
