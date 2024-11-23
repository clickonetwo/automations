/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"html"
)

func LoginForm(message string) []byte {
	head := `
<head>
    <title>Dialpad History Login</title>
	<style>
		body {
			font-family: sans-serif;
			background-color: #f0f0f0;
		}
		.login-container {
			max-width: 60%;
			margin: 100px auto;
			padding: 20px;
			background-color: #ffffff;
			box-shadow: 0px 0px 10px #888888;
		}
		.message {
			color: red;
			text-align: center;
		}
		input[type="text"],
		input[type="password"] {
			width: 80%;
			padding: 10px;
			margin-bottom: 10px;
		}
		button {
			background-color: #4CAF50;
			color: white;
			padding: 10px 20px;
			border: none;
			cursor: pointer;
			width: 100%;
		}
		button:hover {
			background-color: #45a049;
		}
	</style>
</head>`
	bodyTop := `
<body>
    <div class="login-container">
        <h2>Dialpad History Login</h2>`
	messageStart := `<p class="message">`
	messageEnd := `</p>`
	form := `
        <form action="/login" method="GET">
            <label for="username">Username:</label>
            <input type="text" id="username" name="username" required><br>
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required><br>
            <button type="submit">Login</button>
        </form>`
	bodyBottom := `
    </div>
</body>
`
	messageBox := "\n"
	if message != "" {
		messageBox += messageStart + html.EscapeString(message) + messageEnd + "\n"
	}
	page := `<!DOCTYPE html><html>` + head + bodyTop + messageBox + form + bodyBottom + `</html>`
	return []byte(page)
}

func LoginSuccessForm() []byte {
	page := `
<!DOCTYPE html>
<html>
	<head>
		<title>Login successful</title>
		<meta charset="utf-8">
		<meta http-equiv="refresh" content="0;url=/history" />
		<style type="text/css">
			p {
				font-family: sans-serif;
				text-align: center;
			}
		</style>
	</head>
	<body>
		<p>You have successfully logged in!
			<a href="/history">Click here</a> if you are not redirected in a few seconds.</p>
	</body>
</html>`
	return []byte(page)
}
